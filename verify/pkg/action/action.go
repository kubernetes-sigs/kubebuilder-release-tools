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
	"fmt"
	"sync"

	"sigs.k8s.io/kubebuilder-release-tools/verify/pkg/log"
)

// action executes the wrapped plugins concurrently
type action struct {
	plugins []Plugin

	log.Logger
}

// New creates a new Action which will run the provided plugins
func New(plugins ...Plugin) Action {
	return action{
		plugins: plugins,
		Logger: log.New(),
	}
}

// Run implements Action
func (a action) Run() {
	env, err := newPREnv()
	if err != nil {
		a.Fatalf(1, "%v", err)
	}
	a.Debugf("environment for %s/%s ready", env.Owner, env.Repo)

	res := make(chan error)
	var done sync.WaitGroup

	for _, p := range a.plugins {
		// Required to scope it to prevent the use of a loop variable inside a function literal
		plugin := p

		a.Debugf("launching %q plugins", plugin.Name())
		done.Add(1)
		go func() {
			defer done.Done()
			res <- plugin.Entrypoint(env)
		}()
	}

	go func() {
		done.Wait()
		close(res)
	}()

	a.Debug("retrieving plugin results")
	errCount := 0
	for err := range res {
		if err == nil {
			continue
		}
		errCount++
		a.Errorf("%v", err)
	}

	a.Infof("%d plugins ran", len(a.plugins))
	if errCount > 0 {
		a.Fatalf(2, "%v", fmt.Errorf("%d plugins had errors", errCount))
	}
	a.Info("Success!")
}
