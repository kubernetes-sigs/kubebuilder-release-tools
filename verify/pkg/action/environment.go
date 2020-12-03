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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

const (
	envActionsKey    = "GITHUB_ACTIONS"
	envRepositoryKey = "GITHUB_REPOSITORY"
	envEventPathKey  = "GITHUB_EVENT_PATH"
	envTokenKey      = "INPUT_GITHUB_TOKEN"
)

type PREnv struct {
	Owner  string
	Repo   string
	Event  *github.PullRequestEvent
	Client *github.Client
}

func newPREnv() (*PREnv, error) {
	if os.Getenv(envActionsKey) != "true" {
		return nil, fmt.Errorf("not running in an action, bailing.  Set GITHUB_ACTIONS and the other appropriate env vars if you really want to do this.")
	}

	// Get owner and repository
	ownerAndRepo := strings.Split(os.Getenv(envRepositoryKey), "/")

	// Get event path
	eventPath := os.Getenv(envEventPathKey)
	if eventPath == "" {
		return nil, fmt.Errorf("no event path set, something weird is up")
	}

	// Parse the event
	event, err := func() (github.PullRequestEvent, error) {
		eventFile, err := os.Open(eventPath)
		if err != nil {
			return github.PullRequestEvent{}, fmt.Errorf("unable to load event file: %w", err)
		}
		defer func() {
			// As we are not writing to the file, we can omit the error
			_ = eventFile.Close()
		}()

		var event github.PullRequestEvent
		if err := json.NewDecoder(eventFile).Decode(&event); err != nil {
			return event, fmt.Errorf("unable to unmarshal event: %w", err)
		}
		return event, nil
	}()
	if err != nil {
		return nil, err
	}

	// Create the client
	client := github.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv(envTokenKey)},
	)))

	return &PREnv{
		Owner:  ownerAndRepo[0],
		Repo:   ownerAndRepo[1],
		Event:  &event,
		Client: client,
	}, nil
}
