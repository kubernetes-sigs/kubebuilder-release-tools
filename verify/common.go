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

package verify

import (
	"fmt"
	"os"
	"encoding/json"
	"errors"
	"context"
	"time"
	"strings"
	"sync"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

type ErrWithHelp interface {
	error
	Help() string
}

type PRPlugin struct {
	ForAction func(string) bool
	ProcessPR func(pr *github.PullRequest) (string, error)
	Name string
	Title string
}

func (p *PRPlugin) Entrypoint(env *ActionsEnv) error {
	if p.ForAction != nil && !p.ForAction(env.Event.GetAction()) {
		return nil
	}

	repoParts := strings.Split(env.Event.GetRepo().GetFullName(), "/")
	orgName, repoName := repoParts[0], repoParts[1]
	
	headSHA := env.Event.GetPullRequest().GetHead().GetSHA()
	fmt.Printf("::debug::creating check run %q on %s/%s @ %s...\n", p.Name, orgName, repoName, headSHA)

	resRun, runResp, err := env.Client.Checks.CreateCheckRun(context.TODO(), orgName, repoName, github.CreateCheckRunOptions{
		Name: p.Name,
		HeadSHA: headSHA,
		Status: github.String("in_progress"),
	})
	if err != nil {
		return fmt.Errorf("unable to submit check result: %w", err)
	}

	env.Debugf("create check API response: %+v", runResp)
	env.Debugf("created run: %+v", resRun)

	successStatus, procErr := p.ProcessPR(env.Event.PullRequest)

	var summary, fullHelp, conclusion string
	if procErr != nil {
		summary = procErr.Error()
		var helpErr ErrWithHelp
		if errors.As(procErr, &helpErr) {
			fullHelp = helpErr.Help()
		}
		conclusion = "failure"
	} else {
		summary = "Success"
		fullHelp = successStatus
		conclusion = "success"
	}
	completedAt := github.Timestamp{Time: time.Now()}

	// log in case we can't submit the result for some reason
	env.Debugf("plugin result summary: %q", summary)
	env.Debugf("plugin result details: %q", fullHelp)
	env.Debugf("plugin conclusion: %q", conclusion)

	resRun, updateResp, err := env.Client.Checks.UpdateCheckRun(context.TODO(), orgName, repoName, resRun.GetID(), github.UpdateCheckRunOptions{
		Name: p.Name,
		Status: github.String("completed"),
		Conclusion: github.String(conclusion), 
		CompletedAt: &completedAt,
		Output: &github.CheckRunOutput{
			Title: github.String(p.Title),
			Summary: github.String(summary),
			Text: github.String(fullHelp),
		},
	})
	if err != nil {
		return fmt.Errorf("unable to update check result: %w", err)
	}

	env.Debugf("update check API response: %+v", updateResp)
	env.Debugf("updated run: %+v", resRun)

	// return failure here too so that the whole suite fails (since the actions
	// suite seems to ignore failing check runs when calculating general failure)
	if procErr != nil {
		return fmt.Errorf("failed: %v", procErr)
	}
	return nil
}

type ActionsEnv struct {
	Event *github.PullRequestEvent
	Client *github.Client
}
func (ActionsEnv) Errorf(fmtStr string, args ...interface{}) {
	fmt.Printf("::error::"+fmtStr+"\n", args...)
}
func (ActionsEnv) Debugf(fmtStr string, args ...interface{}) {
	fmt.Printf("::debug::"+fmtStr+"\n", args...)
}
func (ActionsEnv) Warnf(fmtStr string, args ...interface{}) {
	fmt.Printf("::warning::"+fmtStr+"\n", args...)
}

func SetupEnv() (*ActionsEnv, error) {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		return nil, fmt.Errorf("not running in an action, bailing.  Set GITHUB_ACTIONS and the other appropriate env vars if you really want to do this.")
	}

	payloadPath := os.Getenv("GITHUB_EVENT_PATH")
	if payloadPath == "" {
		return nil, fmt.Errorf("no payload path set, something weird is up")
	}

	payload, err := func() (github.PullRequestEvent, error) {
		payloadRaw, err := os.Open(payloadPath)
		if err != nil {
			return github.PullRequestEvent{}, fmt.Errorf("unable to load payload file: %w", err)
		}
		defer payloadRaw.Close()
		
		var payload github.PullRequestEvent
		if err := json.NewDecoder(payloadRaw).Decode(&payload); err != nil {
			return payload, fmt.Errorf("unable to unmarshal payload: %w", err) 
		}
		return payload, nil
	}()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	authClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("INPUT_GITHUB_TOKEN")},
	))

	return &ActionsEnv{
		Event: &payload,
		Client: github.NewClient(authClient),
	}, nil
}

type ActionsCallback func(*ActionsEnv) error
func ActionsEntrypoint(cb ActionsCallback) {
	env, err := SetupEnv()
	if err != nil {
		env.Errorf("%v", err)
		os.Exit(1)
	}

	if err := cb(env); err != nil {
		env.Errorf("%v", err)
		os.Exit(2)
	}
	fmt.Println("Success!")
}

func RunPlugins(plugins ...PRPlugin) ActionsCallback {
	return func(env *ActionsEnv) error {
		res := make(chan error)
		var done sync.WaitGroup

		for _, plugin := range plugins {
			done.Add(1)
			go func(plugin PRPlugin) {
				defer done.Done()
				res <- plugin.Entrypoint(env)
			}(plugin)
		}

		go func() {
			done.Wait()
			close(res)
		}()

		errCount := 0
		for err := range res {
			if err == nil {
				continue
			}
			errCount++
			env.Errorf("%v", err)
		}

		fmt.Printf("%d plugins ran\n", len(plugins))
		if errCount > 0 {
			return fmt.Errorf("%d plugins had errors", errCount)
		}
		return nil
	}
}
