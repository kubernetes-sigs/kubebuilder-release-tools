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

package common

import (
	"fmt"
	"strings"
)

type PRType int
func (t PRType) Emoji() string {
	switch t {
	case UncategorizedPR:
		return "<uncategorized>"
	case BreakingPR:
		return emojiBreaking
	case FeaturePR:
		return emojiFeature
	case BugfixPR:
		return emojiBugfix
	case DocsPR:
		return emojiDocs
	case InfraPR:
		return emojiInfra
	default:
		panic(fmt.Sprintf("unrecognized PR type %v", t))
	}
}
func (t PRType) String() string {
	switch t {
	case UncategorizedPR:
		return "uncategorized"
	case BreakingPR:
		return "breaking"
	case FeaturePR:
		return "feature"
	case BugfixPR:
		return "bugfix"
	case DocsPR:
		return "docs"
	case InfraPR:
		return "infra"
	default:
		panic(fmt.Sprintf("unrecognized PR type %v", t))
	}
}

const (
	UncategorizedPR PRType = iota
	BreakingPR
	FeaturePR
	BugfixPR
	DocsPR
	InfraPR
)

// NB(directxman12): These are constants because some folks' dev environments like
// to inject extra combining characters into the mix (generally variation selector 16,
// which indicates emoji presentation), so we want to check that these are *just* the
// character without the combining parts.  Note that they're a rune, so that they
// can *only* be one codepoint.
const (
	emojiFeature     = string('✨')
	emojiBugfix      = string('🐛')
	emojiDocs        = string('📖')
	emojiInfra       = string('🌱')
	emojiBreaking    = string('⚠')
	emojiInfraLegacy = string('🏃')
)

func PRTypeFromTitle(title string) (PRType, string) {
	title = strings.TrimSpace(title)

	if len(title) == 0 {
		return UncategorizedPR, title
	}

	var prType PRType
	switch {
	case strings.HasPrefix(title, ":sparkles:"), strings.HasPrefix(title, emojiFeature):
		title = strings.TrimPrefix(title, ":sparkles:")
		title = strings.TrimPrefix(title, emojiFeature)
		prType = FeaturePR
	case strings.HasPrefix(title, ":bug:"), strings.HasPrefix(title, emojiBugfix):
		title = strings.TrimPrefix(title, ":bug:")
		title = strings.TrimPrefix(title, emojiBugfix)
		prType = BugfixPR
	case strings.HasPrefix(title, ":book:"), strings.HasPrefix(title, emojiDocs):
		title = strings.TrimPrefix(title, ":book:")
		title = strings.TrimPrefix(title, emojiDocs)
		prType = DocsPR
	case strings.HasPrefix(title, ":seedling:"), strings.HasPrefix(title, emojiInfra):
		title = strings.TrimPrefix(title, ":seedling:")
		title = strings.TrimPrefix(title, emojiInfra)
		prType = InfraPR
	case strings.HasPrefix(title, ":warning:"), strings.HasPrefix(title, emojiBreaking):
		title = strings.TrimPrefix(title, ":warning:")
		title = strings.TrimPrefix(title, emojiBreaking)
		prType = BreakingPR
	case strings.HasPrefix(title, ":running:"), strings.HasPrefix(title, emojiInfraLegacy):
		// This has been deprecated in favor of :seedling:
		title = strings.TrimPrefix(title, ":running:")
		title = strings.TrimPrefix(title, emojiInfraLegacy)
		prType = InfraPR
	default:
		return UncategorizedPR, title
	}

	// strip the variation selector from the title, if present
	// (some systems sneak it in -- my guess is OSX)
	title = strings.TrimPrefix(title, "\uFE0F")

	// NB(directxman12): there are a few other cases like the variation selector,
	// but I can't seem to dig them up.  If something doesn't parse as expected,
	// check for zero-width characters and add handling here.

	return prType, strings.TrimSpace(title)
}
