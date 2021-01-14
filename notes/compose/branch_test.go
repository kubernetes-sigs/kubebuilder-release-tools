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
	"fmt"

	"github.com/blang/semver/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "sigs.k8s.io/kubebuilder-release-tools/notes/compose"
	"sigs.k8s.io/kubebuilder-release-tools/notes/pkg/git"
)

var _ = Describe("Branches", func() {
	Describe("finding the latest release", func() {
		branch := ReleaseBranch{Version: semver.Version{Major: 1}}
		It("should return ReleaseTag if there was a release in this branch's history", func() {
			gitImpl := git.UtilitiesMock{
				ClosestTagF: func(git.Committish) (git.Tag, error) {
					return git.Tag("v1.3.4"), nil
				},
			}

			Expect(branch.LatestRelease(gitImpl, false)).To(Equal(ReleaseTag(
				semver.Version{Major: 1, Minor: 3, Patch: 4},
			)))
		})

		It("should support pre-release ReleaseTags", func() {
			gitImpl := git.UtilitiesMock{
				ClosestTagF: func(git.Committish) (git.Tag, error) {
					return git.Tag("v1.3.4-alpha.6"), nil
				},
			}

			Expect(branch.LatestRelease(gitImpl, false)).To(Equal(ReleaseTag(
				semver.Version{Major: 1, Minor: 3, Patch: 4, Pre: []semver.PRVersion{
					{VersionStr: "alpha"},
					{VersionNum: 6, IsNum: true},
				}},
			)))
		})

		It("should return RootCommit if no release exists yet", func() {
			gitImpl := git.UtilitiesMock{
				ClosestTagF: func(git.Committish) (git.Tag, error) {
					return git.Tag(""), fmt.Errorf("no tag found!")
				},
				RootCommitF: func(git.Ref) (git.Commit, error) {
					return git.Commit("abcdef"), nil
				},
			}
			Expect(branch.LatestRelease(gitImpl, false)).To(Equal(FirstCommit{
				Branch: branch,
				Commit: git.Commit("abcdef"),
			}))
		})

		It("should fail if no release exists and the first commit cannot be found", func() {
			gitImpl := git.UtilitiesMock{
				ClosestTagF: func(git.Committish) (git.Tag, error) {
					return git.Tag(""), fmt.Errorf("no tag found!")
				},
				RootCommitF: func(git.Ref) (git.Commit, error) {
					return git.Commit(""), fmt.Errorf("infinite parallel lines, non-euclidean git repository encountered!")
				},
			}
			_, err := branch.LatestRelease(gitImpl, false)
			Expect(err).To(HaveOccurred())
		})

		It("should reject tags from the wrong branch if asked to verify tags", func() {
			gitImpl := git.UtilitiesMock{
				ClosestTagF: func(git.Committish) (git.Tag, error) {
					return git.Tag("v0.6.7"), nil
				},
			}

			_, err := branch.LatestRelease(gitImpl, true)
			Expect(err).To(HaveOccurred())
		})

		It("should accept tags from the wrong branch if not asked to verify tags", func() {
			gitImpl := git.UtilitiesMock{
				ClosestTagF: func(git.Committish) (git.Tag, error) {
					return git.Tag("v0.6.7"), nil
				},
			}

			Expect(branch.LatestRelease(gitImpl, false)).To(Equal(ReleaseTag(
				semver.Version{Minor: 6, Patch: 7},
			)))
		})
	})

	Describe("verifying tags belong to this branch", func() {
		Context("when dealing with release-X branches", func() {
			branch := ReleaseBranch{Version: semver.Version{Major: 2}}
			It("should accept tags with matching X versions", func() {
				tag := ReleaseTag(semver.Version{Major: 2, Minor: 3, Patch: 1})
				Expect(branch.VerifyTagBelongs(tag)).To(Succeed())
			})
			It("should reject tags with different X versions", func() {
				tag := ReleaseTag(semver.Version{Major: 1, Minor: 3, Patch: 1})
				Expect(branch.VerifyTagBelongs(tag)).NotTo(Succeed())
			})
		})
		Context("when dealing with release-0.Y branches", func() {
			branch := ReleaseBranch{Version: semver.Version{Minor: 3}}
			It("should accept tags with matching Y versions", func() {
				tag := ReleaseTag(semver.Version{Major: 0, Minor: 3, Patch: 6})
				Expect(branch.VerifyTagBelongs(tag)).To(Succeed())
			})
			It("should reject tags with different Y versions", func() {
				tag := ReleaseTag(semver.Version{Major: 0, Minor: 4, Patch: 1})
				Expect(branch.VerifyTagBelongs(tag)).NotTo(Succeed())
			})
			It("should reject tags with non-zero X versions", func() {
				// NB: matching X version here
				tag := ReleaseTag(semver.Version{Major: 1, Minor: 3})
				Expect(branch.VerifyTagBelongs(tag)).NotTo(Succeed())
			})
		})
	})

	Describe("creating from a raw branch name", func() {
		It("should accept release-X branches as vX.0.0 versions", func() {
			Expect(ReleaseFromBranch("release-2")).To(Equal(ReleaseBranch{
				Version: semver.Version{Major: 2},
			}))
		})

		It("should accept release-0.Y branches as v0.Y.0 versions", func() {
			Expect(ReleaseFromBranch("release-0.3")).To(Equal(ReleaseBranch{
				Version: semver.Version{Minor: 3},
			}))
		})

		It("should reject release-0", func() {
			_, err := ReleaseFromBranch("release-0")
			Expect(err).To(HaveOccurred())
		})

		It("should reject release-0.0", func() {
			_, err := ReleaseFromBranch("release-0.0")
			Expect(err).To(HaveOccurred())
		})

		It("should reject release-1.Y", func() {
			_, err := ReleaseFromBranch("release-1.3")
			Expect(err).To(HaveOccurred())
		})

		It("should reject branches not starting with release-", func() {
			_, err := ReleaseFromBranch("feature/sheep-shearing-machine")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("when printing/reinterpreting", func() {
		It("should print vX.y.z as release-X", func() {
			branch := ReleaseBranch{Version: semver.Version{Major: 2}}
			Expect(branch.String()).To(Equal("release-2"))
		})

		It("should print v0.Y.z as release-0.Y", func() {
			branch := ReleaseBranch{Version: semver.Version{Minor: 3}}
			Expect(branch.String()).To(Equal("release-0.3"))
		})

		It("should append @{u} if asked to use upstream branches", func() {
			branch := ReleaseBranch{Version: semver.Version{Minor: 3}, UseUpstream: true}
			Expect(branch.String()).To(Equal("release-0.3@{u}"))

			branch = ReleaseBranch{Version: semver.Version{Major: 2}, UseUpstream: true}
			Expect(branch.String()).To(Equal("release-2@{u}"))
		})
	})
})
