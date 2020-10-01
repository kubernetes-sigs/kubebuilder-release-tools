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

package git

import (
	"fmt"
	"os/exec"
	"strings"

	"sigs.k8s.io/kubebuilder-release-tools/notes/common"
)

// Committish represents some git committish object.
type Committish interface {
	Committish() string
}

// Tag is a committish that is a git tag.
type Tag string

func (t Tag) Committish() string {
	return string(t)
}

// Commit is a committish that is a git commit SHA.
type Commit string

func (c Commit) Committish() string {
	return string(c)
}

// SomeCommittish is any ol' user-specified Committish.
type SomeCommittish string

func (c SomeCommittish) Committish() string {
	return string(c)
}

// Git runs the git-related functionality used by the release notes package,
// (that way a mock can be produced).
type Git interface {
	// ClosestTag finds the closest tag to the given committish.
	ClosestTag(initial Committish) (Tag, error)
	// FirstCommit finds the first commit on a given branch.
	FirstCommit(branchName string) (Commit, error)
	// HasUpstream checks if a given branch has an upstream, returning an error if it does not.
	HasUpstream(branchName string) error
	// MergeCommitsBetween shows all the merge commits between start and end,
	// in %B (raw body) form.
	MergeCommitsBetween(start, end Committish) (string, error)
}

// Actual calls out to the git command to get results.
var Actual = actualGit{}

// actualGit calls out to the git command to get results.
type actualGit struct{}

func (actualGit) ClosestTag(initial Committish) (Tag, error) {
	latestTagCmd := exec.Command("git", "describe", "--tags", "--abbrev=0", initial.Committish())
	tagRaw, err := latestTagCmd.Output()
	if err != nil {
		return Tag(""), common.ErrOut(err)
	}

	return Tag(strings.TrimSpace(string(tagRaw))), nil
}
func (actualGit) FirstCommit(branchName string) (Commit, error) {
	cmd := exec.Command("git", "rev-list", "--max-parents=0", branchName)
	out, err := cmd.Output()
	if err != nil {
		return "", common.ErrOut(err)
	}
	return Commit(strings.TrimSpace(string(out))), nil
}
func (actualGit) HasUpstream(branchName string) error {
	return exec.Command("git", "rev-parse", "--abbrev=0", "--symbolic-full-name", branchName).Run()
}
func (actualGit) CurrentBranch() (string, error) {
	currentBranchName, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("unable to determine current branch from HEAD: %w", common.ErrOut(err))
	}
	return strings.TrimSpace(string(currentBranchName)), err
}
func (actualGit) MergeCommitsBetween(start, end Committish) (string, error) {
	listCommits := exec.Command("git", "rev-list", start.Committish()+".."+end.Committish(), "--merges", "--pretty=format:%B")

	commitsRaw, err := listCommits.Output()
	if err != nil {
		return "", err
	}
	return string(commitsRaw), nil
}

// RemoteForUpstreamFor returns the remote for the upstream for the given branch.
func (actualGit) RemoteForUpstreamFor(branchName string) (string, error) {
	remoteForBranch, err := exec.Command("git", "for-each-ref", "--format=%(upstream:remotename)", "refs/heads/"+branchName).Output()
	if err != nil {
		return "", common.ErrOut(err)
	}
	res := strings.TrimSpace(string(remoteForBranch))
	if res == "" {
		return "", fmt.Errorf("no upstream/remote found")
	}
	return res, nil
}

// URLForRemote returns the fetch URL for the given remote.
func (actualGit) URLForRemote(remote string) (string, error) {
	upstreamURLRaw, err := exec.Command("git", "remote", "get-url", remote).Output()
	if err != nil {
		return "", common.ErrOut(err)
	}
	return strings.TrimSpace(string(upstreamURLRaw)), nil
}

// Fetch fetches the given remote (including tags)
func (actualGit) Fetch(remote string) error {
	return common.ErrOut(exec.Command("git", "fetch", "--tags", remote).Run())
}
