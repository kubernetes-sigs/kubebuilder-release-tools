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

package main

import (
	"github.com/google/go-github/v32/github"
	"testing"
)

func stringPointer(s string) *string {
	return &s
}

func Test_verifyPRType(t *testing.T) {
	tests := []struct {
		name string
		pr   *github.PullRequest
		want string
	}{
		{
			name: "Bugfix PR",
			pr: &github.PullRequest{
				Title: stringPointer(":bug: Fixing bug"),
			},
			want: "Found 🐛 PR (bugfix)",
		},
		{
			name: "Release PR",
			pr: &github.PullRequest{
				Title: stringPointer(":rocket: Release v0.0.1"),
			},
			want: "Found 🚀 PR (release)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _, _ := verifyPRType(tt.pr); got != tt.want {
				t.Errorf("verifyPRType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_trimTitle(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{
			name:  "regular PR title",
			title: "📖 book: Use relative links in generate CRDs doc",
			want:  "📖 book: Use relative links in generate CRDs doc",
		},
		{
			name:  "PR title with WIP",
			title: "WIP 📖 book: Use relative links in generate CRDs doc",
			want:  "📖 book: Use relative links in generate CRDs doc",
		},
		{
			name:  "PR title with [WIP]",
			title: "[WIP] 📖 book: Use relative links in generate CRDs doc",
			want:  "📖 book: Use relative links in generate CRDs doc",
		},
		{
			name:  "PR title with [release-1.0]",
			title: "[release-1.0] 📖 book: Use relative links in generate CRDs doc",
			want:  "📖 book: Use relative links in generate CRDs doc",
		},
		{
			name:  "PR title with [WIP][release-1.0]",
			title: "[WIP][release-1.0] 📖 book: Use relative links in generate CRDs doc",
			want:  "📖 book: Use relative links in generate CRDs doc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trimTitle(tt.title); got != tt.want {
				t.Errorf("trimTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}
