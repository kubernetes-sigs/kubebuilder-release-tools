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
	"strings"
)

// Utilities provides a git-related higher abstraction interface.
type Utilities interface {
	// CurrentBranch returns the current active branch
	CurrentBranch() (Branch, error)

	// ClosestTag finds the closest tag to the given committish.
	ClosestTag(Committish) (Tag, error)
	// RootCommit finds the first commit on a given ref.
	RootCommit(Ref) (Commit, error)

	// HasUpstream checks if a given branch has an upstream.
	HasUpstream(LocalBranch) bool
	// RemoteFor returns the remote for the upstream for the given LocalBranch.
	RemoteFor(LocalBranch) (string, error)
	// RefreshUpstream fetches the upstream remote for the given LocalBranch (including tags).
	RefreshUpstream(LocalBranch) error
	// URLForRemote returns the fetch URL for the given remote.
	URLForRemote(string) (string, error)

	// MergeCommitsBetween shows all the merge commits between start and end, in %B (raw body) form.
	MergeCommitsBetween(start, end Committish) (string, error)
}

// Utils calls out to the undelying CLI to get results.
var Utils = utilities{cli: Command}

// utilities calls out to the undelying CLI to get results.
type utilities struct{
	// cli allows to provide a mock CLI for tests.
	cli CLI
}

// CurrentBranch implements Utilities.CurrentBranch.
func (u utilities) CurrentBranch() (Branch, error) {
	currentBranchName, err := u.cli.RevParse(RevParseOptions{
		Committish: Head,
		AbbrevRef:  true, // --abbrev-ref
	})
	if err != nil {
		return "", fmt.Errorf("unable to determine current branch from %s: %w", Head.Committish(), err)
	}

	return Branch(strings.TrimSpace(currentBranchName)), nil
}

// ClosestTag implements Utilities.ClosestTag.
func (u utilities) ClosestTag(initial Committish) (Tag, error) {
	tag, err := u.cli.Describe(DescribeOptions{
		Committish: initial,
		Tags:       true,    // --tags
		Abbrev:     intP(0), // --abbrev=0
	})
	if err != nil {
		return "", fmt.Errorf("unable to determine closest tag of %q: %w", initial.Committish(), err)
	}

	return Tag(strings.TrimSpace(tag)), nil
}

// RootCommit implements Utilities.RootCommit.
func (u utilities) RootCommit(ref Ref) (Commit, error) {
	commit, err := u.cli.RevList(RevListOptions{
		Committish: ref,
		MaxParents: intP(0), // --max-parents=0
	})
	if err != nil {
		return "", fmt.Errorf("unable to determine root commit of %q: %w", ref.Committish(), err)
	}

	return Commit(strings.TrimSpace(commit)), nil
}

// HasUpstream implements Utilities.HasUpstream.
func (u utilities) HasUpstream(branch LocalBranch) bool {
	out, _ := u.cli.ForEachRef(ForEachRefOptions{
		Committish: branch,
		Format:     "%(upstream)",  // --format="%(upstream)"
	})
	return strings.TrimSpace(out) != ""
}

// RemoteFor implements Utilities.RemoteFor.
func (u utilities) RemoteFor(branch LocalBranch) (string, error) {
	remote, err := u.cli.ForEachRef(ForEachRefOptions{
		Committish: branch,
		Format:     "%(upstream:remotename)", // --format=%(upstream:remotename)
	})
	if err != nil {
		return "", fmt.Errorf("unable to determine remote for %q: %w", branch.Committish(), err)
	}

	remote = strings.TrimSpace(remote)
	if remote == "" {
		return "", fmt.Errorf("no upstream/remote set for branch %q", string(branch))
	}

	return remote, nil
}

// RefreshUpstream implements Utilities.RefreshUpstream.
func (u utilities) RefreshUpstream(branch LocalBranch) error {
	remote, err := u.RemoteFor(branch)
	if err != nil {
		return err
	}

	if _, err := u.cli.Fetch(FetchOptions{
		Remote: remote,
		Tags:   true,   // --tags
	}); err != nil {
		return fmt.Errorf("unable to refresh remote %q: %w", remote, err)
	}

	return nil
}

// URLForRemote returns the fetch URL for the given remote.
func (u utilities) URLForRemote(remote string) (string, error) {
	url, err := u.cli.RemoteGetUrl(RemoteGetUrlOptions{Remote: remote})
	if err != nil {
		return "", fmt.Errorf("unable to get URL of remote %q: %w", remote, err)
	}

	return strings.TrimSpace(url), nil
}

// MergeCommitsBetween implements Utilities.MergeCommitsBetween.
func (u utilities) MergeCommitsBetween(start, end Committish) (string, error) {
	commitList, err := u.cli.RevList(RevListOptions{
		Committish: Range{
			start: start,
			end:   end,
		},
		Merges: true,        // --merges
		Pretty: "format:%B", // --pretty=format:%B
	})
	if err != nil {
		return "", fmt.Errorf("unable to get merge commits between %q and %q: %w",
			start.Committish(), end.Committish(), err)
	}

	return commitList, nil
}

func intP(n int) *int {
	return &n
}
