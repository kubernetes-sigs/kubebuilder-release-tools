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
	"regexp"

	"github.com/google/go-github/v32/github"

	notes "sigs.k8s.io/kubebuilder-release-tools/notes/common"
)

var tagRegexp = regexp.MustCompile(`#\d+\b`)

type prIssueInTitleError struct{}

func (e prIssueInTitleError) Error() string {
	return "Your PR has an Issue or PR number in the title."
}
func (e prIssueInTitleError) Details() string {
	return fmt.Sprintf(`The title should just be descriptive.
Issue numbers belong in the PR body as either %#q (if it closes the issue or PR), or something like %#q (if it's just related).`,
		"Fixes #XYZ", "Related to #XYZ",
	)
}

// checkIssueInTitle verifies that the PR title does not contain any Issue or PR tag
func checkIssueInTitle(pr *github.PullRequest) (string, string, error) {
	_, title := notes.PRTypeFromTitle(pr.GetTitle())
	if tagRegexp.MatchString(title) {
		return "", "", prIssueInTitleError{}
	}

	return "Your PR title does not contain any Issue or PR tags", "", nil
}
