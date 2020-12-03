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
	"strings"

	"github.com/google/go-github/v32/github"

	"sigs.k8s.io/kubebuilder-release-tools/verify/pkg/action"
)

type prDescriptivenessError struct{}

func (e prDescriptivenessError) Error() string {
	return "Your PR description is *really* short."
}
func (e prDescriptivenessError) Details() string {
	return `It probably isn't descriptive enough.
You should give a description that highlights both what you're doing it and *why* you're doing it.
Someone reading the PR description without clicking any issue links should be able to roughly understand what's going on.`
}

// checkPRDescriptiveness
func checkPRDescriptiveness(requiredLines int) action.ValidateFunc {
	return func(pr *github.PullRequest) (string, string, error) {
		lineCnt := 0
		for _, line := range strings.Split(pr.GetBody(), "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			lineCnt++
		}
		if lineCnt < requiredLines {
			return "", "", &prDescriptivenessError{}
		}
		return "Your PR looks descriptive enough!", "", nil
	}
}
