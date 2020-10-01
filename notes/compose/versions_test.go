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
	"sigs.k8s.io/kubebuilder-release-tools/notes/git"
)

var _ = Describe("Versions", func() {
	Describe("finding the current one", func() {
		It("should return the latest release on this branch if it matches the branch version", func() {
			gitImpl := gitFuncs{
				closestTag: func(initial git.Committish) (git.Tag, error) {
					return git.Tag("v0.6.3"), nil
				},
			}

			currBranch := ReleaseBranch{Version: semver.Version{Minor: 6}}
			Expect(CurrentVersion(gitImpl, &currBranch)).To(Equal(ReleaseTag(
				semver.Version{Minor: 6, Patch: 3},
			)))
		})

		It("should return the first commit on this branch if no release exists", func() {
			gitImpl := gitFuncs{
				closestTag: func(initial git.Committish) (git.Tag, error) {
					return git.Tag(""), fmt.Errorf("no tag found!")
				},
				firstCommit: func(branchName string) (git.Commit, error) {
					return git.Commit("abcdef"), nil
				},
			}

			currBranch := ReleaseBranch{Version: semver.Version{Minor: 6}}
			Expect(CurrentVersion(gitImpl, &currBranch)).To(Equal(FirstCommit{
				Branch: currBranch,
				Commit: git.Commit("abcdef"),
			}))
		})

		Context("when figuring out if upstreams should be used", func() {
			It("should clear the upstream on the current branch if no upstream exists", func() {
				gitImpl := gitFuncs{
					closestTag: func(initial git.Committish) (git.Tag, error) {
						if initial.Committish() != "release-0.6" {
							return git.Tag(""), fmt.Errorf("supplied branch that was probably an upstream: %v", initial)
						}
						return git.Tag("v0.6.3"), nil
					},
					hasUpstream: func(branchName string) error {
						if branchName == "release-0.6@{u}" {
							return fmt.Errorf("no upstream for this branch")
						}
						return nil
					},
				}

				currBranch := ReleaseBranch{Version: semver.Version{Minor: 6}, UseUpstream: true}
				_, err := CurrentVersion(gitImpl, &currBranch)
				Expect(err).NotTo(HaveOccurred())
				Expect(currBranch.UseUpstream).To(BeFalse())
			})

			It("should keep the upstream around if one does exist", func() {
				gitImpl := gitFuncs{
					closestTag: func(initial git.Committish) (git.Tag, error) {
						if initial.Committish() != "release-0.6@{u}" {
							return git.Tag(""), fmt.Errorf("supplied branch that was not an upstream: %v", initial)
						}
						return git.Tag("v0.6.3"), nil
					},
					hasUpstream: func(branchName string) error {
						if branchName == "release-0.6@{u}" {
							return nil
						}
						return fmt.Errorf("no upstream for this branch")
					},
				}

				currBranch := ReleaseBranch{Version: semver.Version{Minor: 6}, UseUpstream: true}
				_, err := CurrentVersion(gitImpl, &currBranch)
				Expect(err).NotTo(HaveOccurred())
				Expect(currBranch.UseUpstream).To(BeTrue())
			})

			It("should keep using upstreams when looking back a branch, even if the current one lacked one, if we originally asked to", func() {
				gitImpl := gitFuncs{
					closestTag: func(initial git.Committish) (git.Tag, error) {
						switch initial.Committish() {
						case "release-0.6":
							return git.Tag("v0.5.0"), nil
						case "release-0.5@{u}": // still using the upstream for older branches
							return git.Tag("v0.5.7"), nil
						default:
							panic("unexpected branch requested")
						}
					},
					hasUpstream: func(branchName string) error {
						if branchName == "release-0.6@{u}" {
							return fmt.Errorf("no upstream for this branch")
						}
						return nil
					},
				}

				currBranch := ReleaseBranch{Version: semver.Version{Minor: 6}, UseUpstream: true}
				Expect(CurrentVersion(gitImpl, &currBranch)).To(Equal(ReleaseTag(
					semver.Version{Minor: 5, Patch: 7},
				)))
				Expect(currBranch.UseUpstream).To(BeFalse())
			})
		})

		Context("when the latest release belongs to the previous release", func() {
			It("should return the latest release on that release-0.(Y-1) branch", func() {
				gitImpl := gitFuncs{
					closestTag: func(initial git.Committish) (git.Tag, error) {
						switch initial.Committish() {
						case "release-0.7":
							return git.Tag("v0.6.0"), nil
						case "release-0.6":
							return git.Tag("v0.6.3"), nil
						default:
							panic(fmt.Sprintf("got unexpected commit %v for ClosestTag", initial))
						}
					},
				}

				currBranch := ReleaseBranch{Version: semver.Version{Minor: 7}}
				Expect(CurrentVersion(gitImpl, &currBranch)).To(Equal(ReleaseTag(
					semver.Version{Minor: 6, Patch: 3},
				)))
			})
			It("should return the latest release on that release-(X-1) branch", func() {
				gitImpl := gitFuncs{
					closestTag: func(initial git.Committish) (git.Tag, error) {
						switch initial.Committish() {
						case "release-2":
							return git.Tag("v1.0.0"), nil
						case "release-1":
							return git.Tag("v1.9.6"), nil
						default:
							panic(fmt.Sprintf("got unexpected commit %v for ClosestTag", initial))
						}
					},
				}

				currBranch := ReleaseBranch{Version: semver.Version{Major: 2}}
				Expect(CurrentVersion(gitImpl, &currBranch)).To(Equal(ReleaseTag(
					semver.Version{Major: 1, Minor: 9, Patch: 6},
				)))
			})
			It("should return the latest release on that release-0.Y branch if this is a release-1 branch", func() {
				gitImpl := gitFuncs{
					closestTag: func(initial git.Committish) (git.Tag, error) {
						switch initial.Committish() {
						case "release-1":
							return git.Tag("v0.6.0"), nil
						case "release-0.6":
							return git.Tag("v0.6.3"), nil
						default:
							panic(fmt.Sprintf("got unexpected commit %v for ClosestTag", initial))
						}
					},
				}

				currBranch := ReleaseBranch{Version: semver.Version{Major: 1}}
				Expect(CurrentVersion(gitImpl, &currBranch)).To(Equal(ReleaseTag(
					semver.Version{Minor: 6, Patch: 3},
				)))
			})

			It("should fail if the previous branch has a release that doesn't belong to it", func() {
				gitImpl := gitFuncs{
					closestTag: func(initial git.Committish) (git.Tag, error) {
						switch initial.Committish() {
						case "release-1":
							return git.Tag("v0.6.0"), nil
						case "release-0.6":
							return git.Tag("v0.5.0"), nil
						default:
							panic(fmt.Sprintf("got unexpected commit %v for ClosestTag", initial))
						}
					},
				}

				currBranch := ReleaseBranch{Version: semver.Version{Major: 1}}
				_, err := CurrentVersion(gitImpl, &currBranch)
				Expect(err).To(HaveOccurred())
			})
		})

		It("should fail if the latest release doesn't belong to the current or previous release branch", func() {
			gitImpl := gitFuncs{
				closestTag: func(initial git.Committish) (git.Tag, error) {
					return git.Tag("v0.6.0"), nil
				},
			}

			currBranch := ReleaseBranch{Version: semver.Version{Major: 2}}
			_, err := CurrentVersion(gitImpl, &currBranch)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("finding the next one", func() {
		Context("when making final releases", func() {
			// Test mostly with Pre10 because it's the default from the CLI
			relInfo := ReleaseInfo{Kind: ReleaseFinal, Pre10: true}

			Context("from final releases", func() {
				Context("with X>=1 releases", func() {
					current := ReleaseTag(semver.Version{Major: 1, Minor: 6, Patch: 3})

					It("should bump X on breaking changes", func() {
						log := ChangeLog{
							Breaking: []LogEntry{{Title: "something major", PRNumber: "33"}},
							Bugs:     []LogEntry{{Title: "some bugfix", PRNumber: "55"}},
						}
						Expect(log.ExpectedNextVersion(current, relInfo)).To(Equal(ReleaseTag(
							semver.Version{Major: 2},
						)))
					})

					It("should bump Y on new features", func() {
						log := ChangeLog{
							Features: []LogEntry{{Title: "some feature", PRNumber: "44"}},
							Bugs:     []LogEntry{{Title: "some bugfix", PRNumber: "55"}},
						}
						Expect(log.ExpectedNextVersion(current, relInfo)).To(Equal(ReleaseTag(
							semver.Version{Major: 1, Minor: 7},
						)))
					})

					It("should bump Z on anything else", func() {
						log := ChangeLog{
							Bugs: []LogEntry{{Title: "some bugfix", PRNumber: "55"}},
						}
						Expect(log.ExpectedNextVersion(current, relInfo)).To(Equal(ReleaseTag(
							semver.Version{Major: 1, Minor: 6, Patch: 4},
						)))
					})
				})

				Context("with 0.Y releases", func() {
					current := ReleaseTag(semver.Version{Minor: 6, Patch: 3})

					It("should bump Y on breaking changes with Pre10 set to true", func() {
						log := ChangeLog{
							Breaking: []LogEntry{{Title: "something major", PRNumber: "33"}},
							Features: []LogEntry{{Title: "some feature", PRNumber: "44"}},
						}
						Expect(log.ExpectedNextVersion(current, relInfo)).To(Equal(ReleaseTag(
							semver.Version{Minor: 7},
						)))
					})

					It("should bump Y on new features", func() {
						log := ChangeLog{
							Features: []LogEntry{{Title: "some feature", PRNumber: "44"}},
							Bugs:     []LogEntry{{Title: "some bugfix", PRNumber: "55"}},
						}
						Expect(log.ExpectedNextVersion(current, relInfo)).To(Equal(ReleaseTag(
							semver.Version{Minor: 7},
						)))
					})

					It("should bump Z on anything else", func() {
						log := ChangeLog{
							Docs: []LogEntry{{Title: "some doc change", PRNumber: "66"}},
						}
						Expect(log.ExpectedNextVersion(current, relInfo)).To(Equal(ReleaseTag(
							semver.Version{Minor: 6, Patch: 4},
						)))
					})

					It("should bump 0.Y to 1.0.0 if Pre10 is false", func() {
						v1Info := ReleaseInfo{Kind: ReleaseFinal}
						log := ChangeLog{
							Breaking: []LogEntry{{Title: "something major", PRNumber: "33"}},
							Bugs:     []LogEntry{{Title: "some bugfix", PRNumber: "55"}},
						}
						Expect(log.ExpectedNextVersion(current, v1Info)).To(Equal(ReleaseTag(
							semver.Version{Major: 1},
						)))
					})
				})
			})
			Context("from pre-releases", func() {
				current := ReleaseTag(semver.Version{Major: 2, Pre: []semver.PRVersion{
					{VersionStr: "rc"}, {VersionNum: 4, IsNum: true},
				}})
				It("should just clear the pre-release tag, keeping the version", func() {
					log := ChangeLog{
						Breaking: []LogEntry{{Title: "something major", PRNumber: "33"}},
						Bugs:     []LogEntry{{Title: "some bugfix", PRNumber: "55"}},
					}
					Expect(log.ExpectedNextVersion(current, relInfo)).To(Equal(ReleaseTag(
						semver.Version{Major: 2},
					)))
				})
			})
		})

		Context("when making pre-releases", func() {
			// Test mostly with Pre10 because it's the default from the CLI
			relInfo := ReleaseInfo{Kind: ReleaseBeta, Pre10: true}
			betaPre := func(num uint64) []semver.PRVersion {
				return []semver.PRVersion{{VersionStr: "beta"}, {VersionNum: num, IsNum: true}}
			}

			Context("from final releases", func() {
				Context("with X>=1 releases", func() {
					current := ReleaseTag(semver.Version{Major: 1, Minor: 6, Patch: 3})

					It("should bump X on breaking changes", func() {
						log := ChangeLog{
							Breaking: []LogEntry{{Title: "something major", PRNumber: "33"}},
							Bugs:     []LogEntry{{Title: "some bugfix", PRNumber: "55"}},
						}
						Expect(log.ExpectedNextVersion(current, relInfo)).To(Equal(ReleaseTag(
							semver.Version{Major: 2, Pre: betaPre(0)},
						)))
					})

					It("should bump Y on new features", func() {
						log := ChangeLog{
							Features: []LogEntry{{Title: "some feature", PRNumber: "44"}},
							Bugs:     []LogEntry{{Title: "some bugfix", PRNumber: "55"}},
						}
						Expect(log.ExpectedNextVersion(current, relInfo)).To(Equal(ReleaseTag(
							semver.Version{Major: 1, Minor: 7, Pre: betaPre(0)},
						)))
					})

					It("should bump Z on anything else", func() {
						log := ChangeLog{
							Bugs: []LogEntry{{Title: "some bugfix", PRNumber: "55"}},
						}
						Expect(log.ExpectedNextVersion(current, relInfo)).To(Equal(ReleaseTag(
							semver.Version{Major: 1, Minor: 6, Patch: 4, Pre: betaPre(0)},
						)))
					})
				})

				Context("with 0.Y releases", func() {
					current := ReleaseTag(semver.Version{Minor: 6, Patch: 3})

					It("should bump Y on breaking changes with Pre10 set to true", func() {
						log := ChangeLog{
							Breaking: []LogEntry{{Title: "something major", PRNumber: "33"}},
							Features: []LogEntry{{Title: "some feature", PRNumber: "44"}},
						}
						Expect(log.ExpectedNextVersion(current, relInfo)).To(Equal(ReleaseTag(
							semver.Version{Minor: 7, Pre: betaPre(0)},
						)))
					})

					It("should bump Y on new features", func() {
						log := ChangeLog{
							Features: []LogEntry{{Title: "some feature", PRNumber: "44"}},
							Bugs:     []LogEntry{{Title: "some bugfix", PRNumber: "55"}},
						}
						Expect(log.ExpectedNextVersion(current, relInfo)).To(Equal(ReleaseTag(
							semver.Version{Minor: 7, Pre: betaPre(0)},
						)))
					})

					It("should bump Z on anything else", func() {
						log := ChangeLog{
							Docs: []LogEntry{{Title: "some doc change", PRNumber: "66"}},
						}
						Expect(log.ExpectedNextVersion(current, relInfo)).To(Equal(ReleaseTag(
							semver.Version{Minor: 6, Patch: 4, Pre: betaPre(0)},
						)))
					})

					It("should bump 0.Y to 1.0.0 if Pre10 is false", func() {
						v1Info := ReleaseInfo{Kind: ReleaseBeta}
						log := ChangeLog{
							Breaking: []LogEntry{{Title: "something major", PRNumber: "33"}},
							Bugs:     []LogEntry{{Title: "some bugfix", PRNumber: "55"}},
						}
						Expect(log.ExpectedNextVersion(current, v1Info)).To(Equal(ReleaseTag(
							semver.Version{Major: 1, Pre: betaPre(0)},
						)))
					})
				})
			})
			Context("from pre-releases", func() {
				current := ReleaseTag(semver.Version{Major: 2, Pre: []semver.PRVersion{
					{VersionStr: "beta"}, {VersionNum: 0, IsNum: true},
				}})
				Context("with the same kind of pre-release", func() {
					It("should just increment the pre-release number", func() {
						log := ChangeLog{
							Breaking: []LogEntry{{Title: "something major", PRNumber: "33"}},
							Bugs:     []LogEntry{{Title: "some bugfix", PRNumber: "55"}},
						}
						Expect(log.ExpectedNextVersion(current, relInfo)).To(Equal(ReleaseTag(
							semver.Version{Major: 2, Pre: betaPre(1)},
						)))
					})
				})
				Context("with a different kind of pre-release", func() {
					It("should reset the pre-release info to the desired state if it would be an increment", func() {
						rcInfo := ReleaseInfo{Kind: ReleaseCandidate, Pre10: true}
						log := ChangeLog{
							Breaking: []LogEntry{{Title: "something major", PRNumber: "33"}},
							Bugs:     []LogEntry{{Title: "some bugfix", PRNumber: "55"}},
						}
						Expect(log.ExpectedNextVersion(current, rcInfo)).To(Equal(ReleaseTag(
							semver.Version{Major: 2, Pre: []semver.PRVersion{
								{VersionStr: "rc"}, {VersionNum: 0, IsNum: true},
							}},
						)))
					})

					It("should reject trying to return to older pre-release kinds", func() {
						alphaInfo := ReleaseInfo{Kind: ReleaseAlpha, Pre10: true}
						log := ChangeLog{
							Breaking: []LogEntry{{Title: "something major", PRNumber: "33"}},
							Bugs:     []LogEntry{{Title: "some bugfix", PRNumber: "55"}},
						}

						_, err := log.ExpectedNextVersion(current, alphaInfo)
						Expect(err).To(HaveOccurred())
					})
				})
			})
		})
	})
})
