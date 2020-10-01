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

var (
	shortishCommitList = `commit ac380d61764a160b32946e606b0c9ecd2834e3e8
Merge pull request #1165 from vincepri/backpor06-1163

:bug: [0.6] Controller.Watch() should not store watches if already started
commit 29c2e320531ea96428e10e4cca49e48751cf4ce5
Merge pull request #1137 from vincepri/update-jsonpatch490-06

🌱 [0.6] Update json-patch to v4.9.0
`
	shortishChangeLog = ChangeLog{
		Bugs:  []LogEntry{{Title: "[0.6] Controller.Watch() should not store watches if already started", PRNumber: "1165"}},
		Infra: []LogEntry{{Title: "[0.6] Update json-patch to v4.9.0", PRNumber: "1137"}},
	}
)

var _ = Describe("Change Logs", func() {
	It("should be able to just figure out the latest version if we don't ask for a specific one", func() {
		gitImpl := gitFuncs{
			closestTag: func(initial git.Committish) (git.Tag, error) {
				return git.Tag("v0.6.3"), nil
			},
			mergeCommitsBetween: func(start, end git.Committish) (string, error) {
				if start.Committish() != "v0.6.3" || end.Committish() != "release-0.6" {
					return "", fmt.Errorf("couldn't find commits for unexpected range %s..%s", start.Committish(), end.Committish())
				}
				return shortishCommitList, nil
			},
		}
		currBranch := ReleaseBranch{Version: semver.Version{Minor: 6}}

		log, since, err := Changes(gitImpl, &currBranch)
		Expect(err).NotTo(HaveOccurred())
		Expect(since).To(Equal(ReleaseTag(semver.Version{Minor: 6, Patch: 3})))
		Expect(log).NotTo(Equal(ChangeLog{})) // just don't be empty, we'll test other things later
	})

	It("should use the specified start point when we specify one", func() {
		gitImpl := gitFuncs{
			mergeCommitsBetween: func(start, end git.Committish) (string, error) {
				if start.Committish() != "abcdef" || end.Committish() != "release-0.6" {
					return "", fmt.Errorf("couldn't find commits for unexpected range %s..%s", start.Committish(), end.Committish())
				}
				return shortishCommitList, nil
			},
		}
		currBranch := ReleaseBranch{Version: semver.Version{Minor: 6}}

		log, err := ChangesSince(gitImpl, currBranch, git.SomeCommittish("abcdef"))
		Expect(err).NotTo(HaveOccurred())
		Expect(log).NotTo(Equal(ChangeLog{})) // just don't be empty, we'll test other things later
	})

	It("should fail if we can't get the merge commits", func() {
		gitImpl := gitFuncs{
			mergeCommitsBetween: func(start, end git.Committish) (string, error) {
				// note for non-native speakers: "accidentally the X" is a meme-y colloquialism
				return "", fmt.Errorf("couldn't find the commits -- did you accidentally the repository?")
			},
		}
		currBranch := ReleaseBranch{Version: semver.Version{Minor: 6}}

		_, err := ChangesSince(gitImpl, currBranch, git.SomeCommittish("abcdef"))
		Expect(err).To(HaveOccurred())
	})

	It("should turn merge commits into changelog entries", func() {
		gitImpl := gitFuncs{
			mergeCommitsBetween: func(start, end git.Committish) (string, error) {
				return (
				// a decent sampling of different commits -- at least
				// one of each type, but not necessarily one of each indicator.
				// The full range of indicators is tested in common.
				`commit 6af4e7c71d4ca149837d2ed9a33fd8df98ac6103
Merge pull request #1187 from vincepri/go115

:seedling: Update Go mod version to 1.15
commit fdc6658a141b99a3fcb733c8a8000f98e6666f48
Merge pull request #850 from akutz/feature/createOrPatch

✨CreateOrPatch
commit be18097a47bdf9341e31a700cc1c2c23ebb48e42
Merge pull request #1176 from prafull01/multi-apitype

:sparkles: Add error check for multiple apiTypes as reconciliation object
commit 5757a389803ec368126bb1ff046ae3524dacbfcf
Merge pull request #1155 from DirectXMan12/bug/webhook-server-threadsafe

:bug: Ensure that webhook server is thread/start-safe
commit ea6a506eb2b74d17606171d46675da4ec4053c5b
Merge pull request #1075 from alvaroaleman/add

:running: Proposal to extract cluster-specifics out of the Manager
commit 22a2c58a47971ab46c2ff8fab1bf6494632cd1f5
Merge pull request #1160 from daniel-hutao/patch-1

update Builder.Register() 's comment - one or more
commit 20af9010491c4e97a6d77219d8c22db9b99aa491
Merge pull request #1163 from vincepri/watches-controller-bug

🐛 Controller.Watch() should not store watches if already started
commit d6829e9c4db802eb4d5703d22c6cd87e8bbf91da
Merge pull request #1153 from gogolok/fix_typo

:book: Fix typo
commit 4717461d1f66687d3a82288d3131302d64f11389
Merge pull request #1144 from alvaroaleman/default-le-resourcelock

⚠ Change leaderlock from ConfigMap to ConfigMapsLeasesResourceLock
commit be59d6426fe904ea87b348d49503112b8eb5ccef
Merge pull request #1129 from Shpectator/admission-webhooks-status-response

:warning: admission responses with raw Status
`), nil
			},
		}
		currBranch := ReleaseBranch{Version: semver.Version{Minor: 6}}

		log, err := ChangesSince(gitImpl, currBranch, git.SomeCommittish("abcdef"))
		Expect(err).NotTo(HaveOccurred())
		Expect(log).To(Equal(ChangeLog{
			Breaking: []LogEntry{
				{
					PRNumber: "1144",
					Title:    "Change leaderlock from ConfigMap to ConfigMapsLeasesResourceLock",
				},
				{PRNumber: "1129", Title: "admission responses with raw Status"},
			},
			Features: []LogEntry{
				{PRNumber: "850", Title: "CreateOrPatch"},
				{
					PRNumber: "1176",
					Title:    "Add error check for multiple apiTypes as reconciliation object",
				},
			},
			Bugs: []LogEntry{
				{
					PRNumber: "1155",
					Title:    "Ensure that webhook server is thread/start-safe",
				},
				{
					PRNumber: "1163",
					Title:    "Controller.Watch() should not store watches if already started",
				},
			},
			Docs: []LogEntry{
				{PRNumber: "1153", Title: "Fix typo"},
			},
			Infra: []LogEntry{
				{PRNumber: "1187", Title: "Update Go mod version to 1.15"},
				{
					PRNumber: "1075",
					Title:    "Proposal to extract cluster-specifics out of the Manager",
				},
			},
			Uncategorized: []LogEntry{
				{
					PRNumber: "1160",
					Title:    "update Builder.Register() 's comment - one or more",
				},
			},
		}))
	})

	It("should skip non-GitHub merge commits", func() {

		gitImpl := gitFuncs{
			mergeCommitsBetween: func(start, end git.Committish) (string, error) {
				return (
				// a few valid commits mixed with some (real) bad merges from CR
				`commit 06787b6b0e735e5a56fdfbcd8129effaefec3146
Merge branch 'master' of github.com:bharathi-tenneti/controller-runtime

commit fdc6658a141b99a3fcb733c8a8000f98e6666f48
Merge pull request #850 from akutz/feature/createOrPatch

✨CreateOrPatch
commit 334ea25a398a658afac27dc656e9f46893f79c6c
Merge branch 'upstream-master'

commit 3738249414e4c4e8a2dce2e4328ca5dd00283876
Merge branch 'master' into k8s-1.15.3

commit ea6a506eb2b74d17606171d46675da4ec4053c5b
Merge pull request #1075 from alvaroaleman/add

:running: Proposal to extract cluster-specifics out of the Manager
commit be18097a47bdf9341e31a700cc1c2c23ebb48e42
Merge pull request #1176 from prafull01/multi-apitype

:sparkles: Add error check for multiple apiTypes as reconciliation object
commit 5757a389803ec368126bb1ff046ae3524dacbfcf
Merge pull request #1155 from DirectXMan12/bug/webhook-server-threadsafe

:bug: Ensure that webhook server is thread/start-safe
`), nil
			},
		}
		currBranch := ReleaseBranch{Version: semver.Version{Minor: 6}}

		log, err := ChangesSince(gitImpl, currBranch, git.SomeCommittish("abcdef"))
		Expect(err).NotTo(HaveOccurred())
		Expect(log).To(Equal(ChangeLog{
			Features: []LogEntry{
				{PRNumber: "850", Title: "CreateOrPatch"},
				{
					PRNumber: "1176",
					Title:    "Add error check for multiple apiTypes as reconciliation object",
				},
			},
			Bugs: []LogEntry{
				{
					PRNumber: "1155",
					Title:    "Ensure that webhook server is thread/start-safe",
				},
			},
			Infra: []LogEntry{
				{
					PRNumber: "1075",
					Title:    "Proposal to extract cluster-specifics out of the Manager",
				},
			},
		}))
	})
})
