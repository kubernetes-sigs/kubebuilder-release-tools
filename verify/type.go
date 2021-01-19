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

	notes "sigs.k8s.io/kubebuilder-release-tools/notes/pkg/utils"
)

// Extracted from kubernetes/test-infra/prow/plugins/wip/wip-label.go
var wipRegex = regexp.MustCompile(`(?i)^\W?WIP\W`)

type prTitleTypeError struct {
	title string
}

func (e prTitleTypeError) Error() string {
	return "No matching PR type indicator found in title."
}
func (e prTitleTypeError) Details() string {
	return fmt.Sprintf(
		`I saw a title of %#q, which doesn't seem to have any of the acceptable prefixes.

You need to have one of these as the prefix of your PR title:

- Breaking change: ⚠ (%#q)
- Non-breaking feature: ✨ (%#q)
- Patch fix: 🐛 (%#q)
- Docs: 📖 (%#q)
- Infra/Tests/Other: 🌱 (%#q)

More details can be found at [sigs.k8s.io/kubebuilder-release-tools/VERSIONING.md](https://sigs.k8s.io/kubebuilder-release-tools/VERSIONING.md).`,
		e.title, ":warning:", ":sparkles:", ":bug:", ":book:", ":seedling:")
}

// verifyPRType checks that the PR title contains a prefix that defines its type
func verifyPRType(pr *github.PullRequest) (string, string, error) {
	// Remove the WIP prefix if found
	title := wipRegex.ReplaceAllString(pr.GetTitle(), "")

	prType, finalTitle := notes.PRTypeFromTitle(title)
	if prType == notes.UncategorizedPR {
		return "", "", prTitleTypeError{title: title}
	}

	return fmt.Sprintf("Found %s PR (%s)", prType.Emoji(), prType), fmt.Sprintf(`Final title:

	%s
`, finalTitle), nil
}
