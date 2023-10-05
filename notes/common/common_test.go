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

package common_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "sigs.k8s.io/kubebuilder-release-tools/notes/common"
)

func TestCommon(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Common Release Not Parsing Suite")
}

var _ = Describe("PR title parsing", func() {
	DescribeTable("prefix to type",
		func(title string, expectedType PRType, expectedTitle string) {
			prType, finalTitle := PRTypeFromTitle(title)
			Expect(prType).To(Equal(expectedType))
			Expect(finalTitle).To(Equal(expectedTitle))
		},
		Entry("should match breaking from ⚠", "⚠ Change leaderlock from ConfigMap to ConfigMapsLeasesResourceLock", BreakingPR, "Change leaderlock from ConfigMap to ConfigMapsLeasesResourceLock"),
		Entry("should match breaking from :warning:", ":warning: admission responses with raw Status", BreakingPR, "admission responses with raw Status"),
		Entry("should match feature from ✨", "✨CreateOrPatch", FeaturePR, "CreateOrPatch"),
		Entry("should match feature from :sparkles:", ":sparkles: Add error check for multiple apiTypes as reconciliation object", FeaturePR, "Add error check for multiple apiTypes as reconciliation object"),
		Entry("should match bugfix from 🐛", "🐛 Controller.Watch() should not store watches if already started", BugfixPR, "Controller.Watch() should not store watches if already started"),
		Entry("should match bugfix from :bug:", ":bug: Ensure that webhook server is thread/start-safe", BugfixPR, "Ensure that webhook server is thread/start-safe"),
		Entry("should match docs from 📖", "📖 Nit: improve doc string", DocsPR, "Nit: improve doc string"),
		Entry("should match docs from :book:", ":book: Fix typo", DocsPR, "Fix typo"),
		Entry("should match infra from 🌱", "🌱 some infra stuff (couldn't find in log)", InfraPR, "some infra stuff (couldn't find in log)"),
		Entry("should match infra from :seedling:", ":seedling: Update Go mod version to 1.15", InfraPR, "Update Go mod version to 1.15"),
		Entry("should match infra from 🏃(deprecated)", "🏃 hack/setup-envtest.sh: follow-up from #1092", InfraPR, "hack/setup-envtest.sh: follow-up from #1092"),
		Entry("should match infra from :running: (deprecated)", ":running: Proposal to extract cluster-specifics out of the Manager", InfraPR, "Proposal to extract cluster-specifics out of the Manager"),
		Entry("should match release from :rocket:", ":rocket: release v0.0.1", ReleasePR, "release v0.0.1"),
		Entry("should match release from 🚀", "🚀 release v0.0.1", ReleasePR, "release v0.0.1"),
		Entry("should put anything else as uncategorized", "blah blah", UncategorizedPR, "blah blah"),
	)

	It("should strip space from the start and end of the final title", func() {
		prType, title := PRTypeFromTitle(":sparkles:     this is a feature")
		Expect(title).To(Equal("this is a feature"))
		Expect(prType).To(Equal(FeaturePR))
	})

	It("should strip space before considering the prefix", func() {
		prType, title := PRTypeFromTitle("  :sparkles:this is a feature")
		Expect(title).To(Equal("this is a feature"))
		Expect(prType).To(Equal(FeaturePR))
	})

	It("should strip variation selectors from the start of the final title", func() {
		prType, title := PRTypeFromTitle("✨\uFE0FTruly sparkly")
		Expect(prType).To(Equal(FeaturePR))
		Expect(title).To(Equal("Truly sparkly"))
	})

	It("should ingore emoji in the middle of the message", func() {
		prType, title := PRTypeFromTitle("this is not a ✨ feature")
		Expect(title).To(Equal("this is not a ✨ feature"))
		Expect(prType).To(Equal(UncategorizedPR))
	})

	It("should ignore github text->emoji in the middle of the message", func() {
		prType, title := PRTypeFromTitle("this is not a :sparkles: feature")
		Expect(title).To(Equal("this is not a :sparkles: feature"))
		Expect(prType).To(Equal(UncategorizedPR))
	})
})
