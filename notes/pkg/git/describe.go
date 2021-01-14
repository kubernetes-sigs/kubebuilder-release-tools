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

// DescribeOptions allow to provide the options for a `git describe ...` command.
type DescribeOptions struct {
	Committish  Committish
	Dirty       bool
	DirtyStr    string
	Broken      bool
	BrokenStr   string
	All         bool
	Tags        bool
	Contains    bool
	Abbrev      *int
	ExactMatch  bool
	Candidates  *int
	Debug       bool
	Long        bool
	Match       []string
	Exclude     []string
	Always      bool
	FirstParent bool
}

// validate implements cliOptions.validate.
func (opts DescribeOptions) validate() error {
	if opts.Committish != nil && (opts.Dirty || opts.DirtyStr != "") {
		return fmt.Errorf("--dirty is incompatible with committish")
	}
	if opts.Committish != nil && (opts.Broken || opts.BrokenStr != "") {
		return fmt.Errorf("--broken is incompatible with committish")
	}

	return nil
}

// arguments implements cliOptions.arguments.
func (opts DescribeOptions) arguments() (args []string) {
	args = append(args, "describe")

	if opts.Committish != nil {
		args = append(args, opts.Committish.Committish())
	} else {
		if opts.DirtyStr != "" {
			args = append(args, fmt.Sprintf("--dirty=%q", opts.DirtyStr))
		} else if opts.Dirty {
			args = append(args, "--dirty")
		}
		if opts.BrokenStr != "" {
			args = append(args, fmt.Sprintf("--broken=%q", opts.BrokenStr))
		} else if opts.Broken {
			args = append(args, "--broken")
		}
	}
	if opts.All {
		args = append(args, "--all")
	} else if opts.Tags {
		args = append(args, "--tags")
	}
	if opts.Contains {
		args = append(args, "--contains")
	}
	if opts.Abbrev != nil {
		args = append(args, fmt.Sprintf("--abbrev=%d", *opts.Abbrev))
	}
	if opts.ExactMatch {
		args = append(args, "--exact-match")
	} else if opts.Candidates != nil {
		args = append(args, fmt.Sprintf("--candidates=%d", *opts.Candidates))
	}
	if opts.Debug {
		args = append(args, "--debug")
	}
	if opts.Long {
		args = append(args, "--long")
	}
	for _, match := range opts.Match {
		args = append(args, "--match", match)
	}
	for _, exclude := range opts.Exclude {
		args = append(args, "--exclude", exclude)
	}
	if opts.Always {
		args = append(args, "--always")
	}
	if opts.FirstParent {
		args = append(args, "--first-parent")
	}

	return
}
