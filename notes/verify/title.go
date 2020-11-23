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
	"regexp"

	"sigs.k8s.io/kubebuilder-release-tools/notes/common"
)

// Extracted from kubernetes/test-infra/prow/plugins/wip/wip-label.go
var wipRegex = regexp.MustCompile(`(?i)^\W?WIP\W`)

type prTitleError struct {
	title string
}
func (e *prTitleError) Error() string {
	return "no matching PR type indicator found in title"
}
func (e *prTitleError) Help() string {
	return fmt.Sprintf(
`I saw a title of %[2]s%[1]s%[2]s, which doesn't seem to have any of the acceptable prefixes.

You need to have one of these as the prefix of your PR title:

- Breaking change: ⚠ (%[2]s:warning:%[2]s)
- Non-breaking feature: ✨ (%[2]s:sparkles:%[2]s)
- Patch fix: 🐛 (%[2]s:bug:%[2]s)
- Docs: 📖 (%[2]s:book:%[2]s)
- Infra/Tests/Other: 🌱 (%[2]s:seedling:%[2]s)

More details can be found at [sigs.k8s.io/controller-runtime/VERSIONING.md](https://sigs.k8s.io/controller-runtime/VERSIONING.md).`, e.title, "`")
}

// VerifyPRTitle checks that the PR title matches a valid PR type prefix,
// returning a message describing what was found on success, and a special
// error (with more detailed help via .Help) on failure.
func VerifyPRTitle(title string) (string, error) {
	// Remove the WIP prefix if found
	title = wipRegex.ReplaceAllString(title, "")

	prType, finalTitle := common.PRTypeFromTitle(title)
	if prType == common.UncategorizedPR {
		return "", &prTitleError{title: title}
	}

	return fmt.Sprintf(
`Found %s PR (%s) with final title:

	%s
`, prType.Emoji(), prType, finalTitle), nil
}
