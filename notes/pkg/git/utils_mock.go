/*
Copyright 2021 The Kubernetes Authors.

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

// Ensure that UtilitiesMock implements Utilities at compile time.
var _ Utilities = UtilitiesMock{}

// UtilitiesMock implements Utilities using the functions provided in the fields instead of the actual cli.
// It is meant to allow unit tests of packages that make use of the Utilities interface.
type UtilitiesMock struct {
	CurrentBranchF func() (Branch, error)

	ClosestTagF func(Committish) (Tag, error)
	RootCommitF func(Ref) (Commit, error)

	HasUpstreamF     func(LocalBranch) bool
	RemoteForF       func(LocalBranch) (string, error)
	RefreshUpstreamF func(LocalBranch) error
	UrlForRemoteF    func(string) (string, error)

	MergeCommitsBetweenF func(start, end Committish) (string, error)
}

// UtilitiesMock.CurrentBranch implements Utilities.CurrentBranch.
func (f UtilitiesMock) CurrentBranch() (Branch, error) {
	if f.CurrentBranchF == nil {
		panic("CurrentBranch not expected")
	}
	return f.CurrentBranchF()
}

// UtilitiesMock.ClosestTag implements Utilities.ClosestTag.
func (f UtilitiesMock) ClosestTag(initial Committish) (Tag, error) {
	if f.ClosestTagF == nil {
		panic("ClosestTag not expected")
	}
	return f.ClosestTagF(initial)
}

// UtilitiesMock.RootCommit implements Utilities.RootCommit.
func (f UtilitiesMock) RootCommit(ref Ref) (Commit, error) {
	if f.RootCommitF == nil {
		panic("RootCommit not expected")
	}
	return f.RootCommitF(ref)
}

// UtilitiesMock.HasUpstream implements Utilities.HasUpstream.
func (f UtilitiesMock) HasUpstream(ref LocalBranch) bool {
	if f.HasUpstreamF == nil {
		panic("HasUpstream not expected")
	}
	return f.HasUpstreamF(ref)
}

// UtilitiesMock.RemoteFor implements Utilities.RemoteFor.
func (f UtilitiesMock) RemoteFor(branch LocalBranch) (string, error) {
	if f.RemoteForF == nil {
		panic("RemoteFor not expected")
	}
	return f.RemoteForF(branch)
}

// UtilitiesMock.RefreshUpstream implements Utilities.RefreshUpstream.
func (f UtilitiesMock) RefreshUpstream(branch LocalBranch) error {
	if f.RefreshUpstreamF == nil {
		panic("RefreshUpstream not expected")
	}
	return f.RefreshUpstreamF(branch)
}

// UtilitiesMock.URLForRemote implements Utilities.URLForRemote.
func (f UtilitiesMock) URLForRemote(remote string) (string, error) {
	if f.UrlForRemoteF == nil {
		panic("URLForRemote not expected")
	}
	return f.UrlForRemoteF(remote)
}

// UtilitiesMock.FetchTags implements git.Utilities.FetchTags.
func (f UtilitiesMock) MergeCommitsBetween(start, end Committish) (string, error) {
	if f.MergeCommitsBetweenF == nil {
		panic("MergeCommitsBetween not expected")
	}
	return f.MergeCommitsBetweenF(start, end)
}
