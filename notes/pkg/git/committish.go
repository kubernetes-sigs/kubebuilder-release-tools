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

import (
	"fmt"
)

// Committish represents some git committish object.
type Committish interface {
	Committish() string
}

// head is a committish that represents the current active head.
type head struct{}

// head.Committish implements Committish.Committish.
func (h head) Committish() string {
	return "HEAD"
}

// Head is a Committish that represents the current active head.
var Head = head{}

// Commit is a Committish that represents a git commit.
type Commit string

// Commit.Committish implements Committish.Committish.
func (c Commit) Committish() string {
	return string(c)
}

// Range is a Committish that represents the range between two other committishes.
type Range struct{
	start Committish
	end   Committish
}

// Range.Committish implements Committish.Committish.
func (r Range) Committish() string {
	return fmt.Sprintf("%s..%s", r.start.Committish(), r.end.Committish())
}

// SomeCommittish is any ol' user-specified Committish.
type SomeCommittish string

// SomeCommittish.Committish implements Committish.Committish.
func (c SomeCommittish) Committish() string {
	return string(c)
}

// Ref represents a git branch or tag.
type Ref interface {
	Committish
	isGitRef()
}

// Tag is a Ref that represents a git tag.
type Tag string

// Tag.Committish implements Committish.Committish.
func (t Tag) Committish() string {
	return string(t)
}

// Tag.isGitRef implements Ref.isGitRef.
func (Tag) isGitRef() {}

// Branch is a Ref that represents a git branch.
type Branch string

// Branch.Committish implements Committish.Committish.
func (b Branch) Committish() string {
	return string(b)
}

// Branch.isGitRef implements Ref.isGitRef.
func (Branch) isGitRef() {}

// LocalBranch is a Ref that represents a git branch.
type LocalBranch string

// LocalBranch.Committish implements Committish.Committish.
func (b LocalBranch) Committish() string {
	return "refs/heads/" + string(b)
}

// LocalBranch.isGitRef implements Ref.isGitRef.
func (LocalBranch) isGitRef() {}
