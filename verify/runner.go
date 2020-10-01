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

package main

import (
	"fmt"
	"os"
	"encoding/json"
	"errors"
	"context"
	"time"
	"strings"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"

	"sigs.k8s.io/kubebuilder-release-tools/notes/verify"
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

func (p *PRPlugin) Entrypoint() error {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		return fmt.Errorf("not running in an action, bailing.  Set GITHUB_ACTIONS and the other appropriate env vars if you really want to do this.")
	}

	payloadPath := os.Getenv("GITHUB_EVENT_PATH")
	if payloadPath == "" {
		return fmt.Errorf("no payload path set, something weird is up")
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
		return err
	}

	if p.ForAction != nil && payload.Action != nil && !p.ForAction(*payload.Action) {
		return nil
	}

	successStatus, procErr := p.ProcessPR(payload.PullRequest)

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

	ctx := context.Background() // TODO: timeouts

	authClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("INPUT_GITHUB_TOKEN")},
	))

	client := github.NewClient(authClient)

	repoParts := strings.Split(*payload.Repo.FullName, "/")
	orgName, repoName := repoParts[0], repoParts[1]
	
	headSHA := payload.GetPullRequest().GetHead().GetSHA()
	fmt.Printf("::debug::creating check run %q on %s/%s @ %s...\n", p.Name, orgName, repoName, headSHA)

	// TODO: create before running the plugin, update after
	resRun, runResp, err := client.Checks.CreateCheckRun(ctx, orgName, repoName, github.CreateCheckRunOptions{
		Name: p.Name,
		HeadSHA: headSHA,
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
		return fmt.Errorf("unable to submit check result: %w")
	}

	fmt.Printf("::debug::create response: %+v\n", runResp)
	fmt.Printf("::debug::created run: %+v\n", resRun)


	// as long as the check result upload succeeded, consider this action as a
	// success, and rely on the check result to indicate otherwise.
	return nil
}

func main() {
	plugin := PRPlugin{
		Name: "pr-type-title",
		Title: "PR Type Title Check",
		ProcessPR: func(pr *github.PullRequest) (string, error) {
			return verify.VerifyPRTitle(pr.GetTitle())
		},
		ForAction: func(action string) bool {
			switch action {
			case "opened", "edited", "reopened":
				return true
			default:
				return false
			}
		},
	}

	if err := plugin.Entrypoint(); err != nil {
		fmt.Printf("::error::%v\n", err)
		os.Exit(1)
	}

	fmt.Println("Success!")
}
