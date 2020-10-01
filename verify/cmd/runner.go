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
	"strings"
	"regexp"

	"github.com/google/go-github/v32/github"

	notes "sigs.k8s.io/kubebuilder-release-tools/notes/common"
	notesver "sigs.k8s.io/kubebuilder-release-tools/notes/verify"
	"sigs.k8s.io/kubebuilder-release-tools/verify"
)

type prErrs struct {
	errs []string
}
func (e prErrs) Error() string {
	return fmt.Sprintf("%d issues found with your PR description", len(e.errs))
}
func (e prErrs) Help() string {
	res := make([]string, len(e.errs))
	for _, err := range e.errs {
		parts := strings.Split(err, "\n")
		for i, part := range parts[1:] {
			parts[i+1] = "  "+part
		}
		res = append(res, "- "+strings.Join(parts, "\n"))
	}
	return strings.Join(res, "\n")
}

func main() {
	verify.ActionsEntrypoint(verify.RunPlugins(
		verify.PRPlugin{
			Name: "PR Type",
			Title: "PR Type in Title",
			ProcessPR: func(pr *github.PullRequest) (string, error) {
				return notesver.VerifyPRTitle(pr.GetTitle())
			},
			ForAction: func(action string) bool {
				switch action {
				case "opened", "edited", "reopened":
					return true
				default:
					return false
				}
			},
		},

		verify.PRPlugin{
			Name: "PR Desc",
			Title: "Basic PR Descriptiveness Check",
			ProcessPR: func(pr *github.PullRequest) (string, error) {
				var errs []string
				// TODO(directxman12): add warnings when we have them

				lineCnt := 0
				for _, line := range strings.Split(pr.GetBody(), "\n") {
					if strings.TrimSpace(line) == "" {
						continue
					}
					lineCnt++
				}
				if lineCnt < 2 {
					errs = append(errs, "**your PR body is *really* short**.\n\nIt probably isn't descriptive enough.\nYou should give a description that highlights both what you're doing it and *why* you're doing it. Someone reading the PR description without clicking any issue links should be able to roughly understand what's going on")
				}

				_, title := notes.PRTypeFromTitle(pr.GetTitle())
				if regexp.MustCompile(`#\d{1,}\b`).MatchString(title) {
					errs = append(errs, "**Your PR has an issue number in the title.**\n\nThe title should just be descriptive.\nIssue numbers belong in the PR body as either `Fixes #XYZ` (if it closes the issue or PR), or something like `Related to #XYZ` (if it's just related).")
				}

				if len(errs) == 0 {
					return "Your PR description looks okay!", nil
				}
				return "", prErrs{errs: errs}
			},
			ForAction: func(action string) bool {
				switch action {
				case "opened", "edited", "reopened":
					return true
				default:
					return false
				}
			},
		},
	))
}
