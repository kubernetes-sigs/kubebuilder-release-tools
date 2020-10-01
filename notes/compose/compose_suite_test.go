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

package compose_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/kubebuilder-release-tools/notes/git"
)

func TestCompose(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compose Suite")
}

type gitFuncs struct {
	closestTag           func(initial git.Committish) (git.Tag, error)
	firstCommit          func(branchName string) (git.Commit, error)
	hasUpstream          func(branchName string) error
	mergeCommitsBetween  func(start, end git.Committish) (string, error)
	remoteForUpstreamFor func(branchName string) (string, error)
	urlForRemote         func(remote string) (string, error)
}

func (f gitFuncs) ClosestTag(initial git.Committish) (git.Tag, error) {
	if f.closestTag == nil {
		panic("ClosestTag not expected")
	}
	return f.closestTag(initial)
}
func (f gitFuncs) FirstCommit(branchName string) (git.Commit, error) {
	if f.firstCommit == nil {
		panic("FirstCommit not expected")
	}
	return f.firstCommit(branchName)
}
func (f gitFuncs) HasUpstream(branchName string) error {
	if f.hasUpstream == nil {
		panic("HasUpstream not expected")
	}
	return f.hasUpstream(branchName)
}
func (f gitFuncs) MergeCommitsBetween(start, end git.Committish) (string, error) {
	if f.mergeCommitsBetween == nil {
		panic("MergeCommitsBetween not expected")
	}
	return f.mergeCommitsBetween(start, end)
}
