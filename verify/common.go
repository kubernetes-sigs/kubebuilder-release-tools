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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

var log logger

type ActionsEnv struct {
	Owner  string
	Repo   string
	Event  *github.PullRequestEvent
	Client *github.Client
	Suffix string
}

func setupEnv() (*ActionsEnv, error) {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		return nil, fmt.Errorf("not running in an action, bailing.  Set GITHUB_ACTIONS and the other appropriate env vars if you really want to do this.")
	}

	// Get owner and repository
	ownerAndRepo := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")

	// Get event path
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return nil, fmt.Errorf("no event path set, something weird is up")
	}

	// Parse the event
	event, err := func() (github.PullRequestEvent, error) {
		eventFile, err := os.Open(eventPath)
		if err != nil {
			return github.PullRequestEvent{}, fmt.Errorf("unable to load event file: %w", err)
		}
		defer eventFile.Close()

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
		&oauth2.Token{AccessToken: os.Getenv("INPUT_GITHUB_TOKEN")},
	)))

	return &ActionsEnv{
		Owner:  ownerAndRepo[0],
		Repo:   ownerAndRepo[1],
		Event:  &event,
		Client: client,
		Suffix: os.Getenv("INPUT_SUFFIX"),
	}, nil
}

type ActionsCallback func(*ActionsEnv) error

func ActionsEntrypoint(cb ActionsCallback) {
	env, err := setupEnv()
	if err != nil {
		log.errorf("%v", err)
		os.Exit(1)
	}

	if err := cb(env); err != nil {
		log.errorf("%v", err)
		os.Exit(2)
	}
	fmt.Println("Success!")
}

func RunPlugins(plugins ...PRPlugin) ActionsCallback {
	return func(env *ActionsEnv) error {
		res := make(chan error)
		var done sync.WaitGroup

		for _, plugin := range plugins {
			if env.Suffix != "" {
				plugin.Name += " " + env.Suffix
			}

			done.Add(1)
			go func(plugin PRPlugin) {
				defer done.Done()
				res <- plugin.entrypoint(env)
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
			log.errorf("%v", err)
		}

		fmt.Printf("%d plugins ran\n", len(plugins))
		if errCount > 0 {
			return fmt.Errorf("%d plugins had errors", errCount)
		}
		return nil
	}
}
