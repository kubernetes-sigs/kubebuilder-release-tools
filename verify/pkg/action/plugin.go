/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package action

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/go-github/v32/github"

	"sigs.k8s.io/kubebuilder-release-tools/verify/pkg/log"
)

const (
	actionOpen   = "opened"
	actionReopen = "reopened"
	actionEdit   = "edited"
	actionSync   = "synchronize"
)

// ValidateFunc is the type of the callback that a Plugin will use to validate the PR contents
type ValidateFunc func(*github.PullRequest) (string, string, error)

// plugin performs the wrapped validate and uploads the results using GitHub Check API
type plugin struct {
	checkRunName        string
	checkRunOutputTitle string

	validate ValidateFunc

	log.Logger
}

// New creates a new Plugin that validates a PR event uploading the results
// using GitHub Check API with the provided name and output title.
func NewPlugin(name, title string, validate ValidateFunc) Plugin {
	return plugin{
		checkRunName:        name,
		checkRunOutputTitle: title,
		validate:            validate,
		Logger:              log.NewFor(name),
	}
}

// Name implements Plugin interface.
func (p plugin) Name() string {
	return p.checkRunName
}

// Entrypoint implements Plugin interface.
func (p plugin) Entrypoint(env *PREnv) error {
	switch env.Event.GetAction() {
	case actionOpen:
		return p.onOpen(env)
	case actionReopen:
		return p.onReopen(env)
	case actionEdit:
		return p.onEdit(env)
	case actionSync:
		return p.onSync(env)
	default:
		p.Warningf("action %q received with no defined procedure, skipping", env.Event.GetAction())
	}

	return nil
}

// onOpen handles "opened" actions
func (p plugin) onOpen(env *PREnv) error {
	p.Debugf("%q handler", actionOpen)
	// Create the check run
	checkRun, err := p.createCheckRun(env.Client, env.Owner, env.Repo, env.Event.GetPullRequest().GetHead().GetSHA())
	if err != nil {
		return err
	}

	// Process the PR and submit the results
	_, err = p.validateAndSubmit(env, checkRun)
	return err
}

// onReopen handles "reopened" actions
func (p plugin) onReopen(env *PREnv) error {
	p.Debugf("%q handler", actionReopen)
	// Get the check run
	checkRun, err := p.getCheckRun(env.Client, env.Owner, env.Repo, env.Event.GetPullRequest().GetHead().GetSHA())
	if err != nil {
		return err
	}

	// Rerun the tests if they weren't finished
	if !Finished.Equal(checkRun.GetStatus()) {
		// Process the PR and submit the results
		_, err = p.validateAndSubmit(env, checkRun)
		return err
	}

	// Return failure here too so that the whole suite fails (since the actions
	// suite seems to ignore failing check runs when calculating general failure)
	if *checkRun.Conclusion == "failure" {
		return fmt.Errorf("failed: %v", *checkRun.Output.Summary)
	}
	return nil
}

// onEdit handles "edited" actions
func (p plugin) onEdit(env *PREnv) error {
	p.Debugf("%q handler", actionEdit)
	// Reset the check run
	checkRun, err := p.resetCheckRun(env.Client, env.Owner, env.Repo, env.Event.GetPullRequest().GetHead().GetSHA())
	if err != nil {
		return err
	}

	// Process the PR and submit the results
	_, err = p.validateAndSubmit(env, checkRun)
	return err
}

// onSync handles "synchronize" actions
func (p plugin) onSync(env *PREnv) error {
	p.Debugf("%q handler", actionSync)
	// Get the check run
	checkRun, err := p.getCheckRun(env.Client, env.Owner, env.Repo, env.Event.GetBefore())
	if err != nil {
		return err
	}

	// Rerun the tests if they weren't finished
	if !Finished.Equal(checkRun.GetStatus()) {
		// Process the PR and submit the results
		checkRun, err = p.validateAndSubmit(env, checkRun)
		if err != nil {
			return err
		}
	}

	// Create a duplicate for the new commit
	checkRun, err = p.duplicateCheckRun(env.Client, env.Owner, env.Repo, env.Event.GetAfter(), checkRun)
	if err != nil {
		return err
	}

	// Return failure here too so that the whole suite fails (since the actions
	// suite seems to ignore failing check runs when calculating general failure)
	if *checkRun.Conclusion == "failure" {
		return fmt.Errorf("failed: %v", *checkRun.Output.Summary)
	}
	return nil
}

// validatePR executes the provided validating function and parses the result
func (p plugin) validatePR(pr *github.PullRequest) (conclusion, summary, text string, err error) {
	p.Debug("execute the plugin checks")
	summary, text, err = p.validate(pr)

	if err == nil {
		conclusion = "success"
	} else {
		conclusion = "failure"
		summary = err.Error()
		var detailedErr ErrorWithDetails
		if errors.As(err, &detailedErr) {
			text = detailedErr.Details()
		}
	}

	// Log in case we can't submit the result for some reason
	p.Debugf("plugin conclusion: %q", conclusion)
	p.Debugf("plugin result summary: %q", summary)
	p.Debugf("plugin result details: %q", text)

	return conclusion, summary, text, err
}

// validateAndSubmit performs the validation and submits the result
func (p plugin) validateAndSubmit(env *PREnv, checkRun *github.CheckRun) (*github.CheckRun, error) {
	// Validate the PR
	conclusion, summary, text, validateErr := p.validatePR(env.Event.PullRequest)

	// Update the check run
	checkRun, err := p.finishCheckRun(env.Client, env.Owner, env.Repo, checkRun.GetID(), conclusion, summary, text)
	if err != nil {
		return checkRun, err
	}

	// Return failure here too so that the whole suite fails (since the actions
	// suite seems to ignore failing check runs when calculating general failure)
	if validateErr != nil {
		return checkRun, fmt.Errorf("failed: %v", validateErr)
	}
	return checkRun, nil
}

////////////////////////////////////////////////////////////////////////////////
//                               Check API calls                              //
////////////////////////////////////////////////////////////////////////////////

// createCheckRun creates a new Check-Run.
// It returns an error in case it couldn't be created.
func (p plugin) createCheckRun(client *github.Client, owner, repo, headSHA string) (*github.CheckRun, error) {
	p.Debugf("creating check run %q on %s/%s @ %s...", p.checkRunName, owner, repo, headSHA)

	checkRun, res, err := client.Checks.CreateCheckRun(
		context.TODO(),
		owner,
		repo,
		github.CreateCheckRunOptions{
			Name:    p.checkRunName,
			HeadSHA: headSHA,
			Status:  Started.StringP(),
		},
	)

	p.Debugf("create check API response: %+v", res)
	p.Debugf("created run: %+v", checkRun)

	if err != nil {
		return nil, fmt.Errorf("unable to create check run: %w", err)
	}
	return checkRun, nil
}

// getCheckRun returns the Check-Run, creating it if it doesn't exist.
// It returns an error in case it didn't exist and couldn't be created, or if there are multiple matches.
func (p plugin) getCheckRun(client *github.Client, owner, repo, headSHA string) (*github.CheckRun, error) {
	p.Debugf("getting check run %q on %s/%s @ %s...", p.checkRunName, owner, repo, headSHA)

	checkRunList, res, err := client.Checks.ListCheckRunsForRef(
		context.TODO(),
		owner,
		repo,
		headSHA,
		&github.ListCheckRunsOptions{
			CheckName: github.String(p.checkRunName),
		},
	)

	p.Debugf("list check API response: %+v", res)
	p.Debugf("listed runs: %+v", checkRunList)

	if err != nil {
		return nil, fmt.Errorf("unable to get check run: %w", err)
	}

	switch n := *checkRunList.Total; {
	case n == 0:
		return p.createCheckRun(client, owner, repo, headSHA)
	case n == 1:
		return checkRunList.CheckRuns[0], nil
	case n > 1:
		return nil, fmt.Errorf("multiple instances of `%s` check run found on %s/%s @ %s",
			p.checkRunName, owner, repo, headSHA)
	default: // Should never happen
		return nil, fmt.Errorf("negative number of instances (%d) of `%s` check run found on %s/%s @ %s",
			n, p.checkRunName, owner, repo, headSHA)
	}
}

// resetCheckRun returns the Check-Run with executing status, creating it if it doesn't exist.
// It returns an error in case it didn't exist and couldn't be created, if there are multiple matches,
// or if it exists but couldn't be updated.
func (p plugin) resetCheckRun(client *github.Client, owner, repo string, headSHA string) (*github.CheckRun, error) {
	checkRun, err := p.getCheckRun(client, owner, repo, headSHA)
	// If it errored, or it was created but not finished, we don't need to update it
	if err != nil || Started.Equal(checkRun.GetStatus()) {
		return checkRun, err
	}

	p.Debugf("resetting check run %q on %s/%s...", p.checkRunName, owner, repo)

	checkRun, updateResp, err := client.Checks.UpdateCheckRun(
		context.TODO(),
		owner,
		repo,
		checkRun.GetID(),
		github.UpdateCheckRunOptions{
			Name:   p.checkRunName,
			Status: Started.StringP(),
		},
	)

	p.Debugf("update check API response: %+v", updateResp)
	p.Debugf("updated run: %+v", checkRun)

	if err != nil {
		return checkRun, fmt.Errorf("unable to reset check run: %w", err)
	}
	return checkRun, nil
}

// finishCheckRun updates the Check-Run with id checkRunID setting its output.
// It returns an error in case it couldn't be updated.
func (p plugin) finishCheckRun(client *github.Client, owner, repo string, checkRunID int64, conclusion, summary, text string) (*github.CheckRun, error) {
	p.Debugf("adding results to check run %q on %s/%s...", p.checkRunName, owner, repo)

	// CheckRun.Output.Text is optional, so empty text strings should actually be nil pointers
	var testPointer *string
	if text != "" {
		testPointer = github.String(text)
	}
	checkRun, updateResp, err := client.Checks.UpdateCheckRun(context.TODO(), owner, repo, checkRunID, github.UpdateCheckRunOptions{
		Name:        p.checkRunName,
		Conclusion:  github.String(conclusion),
		CompletedAt: &github.Timestamp{Time: time.Now()},
		Output: &github.CheckRunOutput{
			Title:   github.String(p.checkRunOutputTitle),
			Summary: github.String(summary),
			Text:    testPointer,
		},
	})

	p.Debugf("update check API response: %+v", updateResp)
	p.Debugf("updated run: %+v", checkRun)

	if err != nil {
		return checkRun, fmt.Errorf("unable to update check run with results: %w", err)
	}
	return checkRun, nil
}

// duplicateCheckRun creates a new Check-Run with the same info as the provided one but for a new headSHA
func (p plugin) duplicateCheckRun(client *github.Client, owner, repo, headSHA string, checkRun *github.CheckRun) (*github.CheckRun, error) {
	p.Debugf("duplicating check run %q on %s/%s @ %s...", p.checkRunName, owner, repo, headSHA)

	checkRun, res, err := client.Checks.CreateCheckRun(
		context.TODO(),
		owner,
		repo,
		github.CreateCheckRunOptions{
			Name:        p.checkRunName,
			HeadSHA:     headSHA,
			DetailsURL:  checkRun.DetailsURL,
			ExternalID:  checkRun.ExternalID,
			Status:      checkRun.Status,
			Conclusion:  checkRun.Conclusion,
			StartedAt:   checkRun.StartedAt,
			CompletedAt: checkRun.CompletedAt,
			Output:      checkRun.Output,
		},
	)

	p.Debugf("create check API response: %+v", res)
	p.Debugf("created run: %+v", checkRun)

	if err != nil {
		return checkRun, fmt.Errorf("unable to duplicate check run: %w", err)
	}
	return checkRun, nil
}
