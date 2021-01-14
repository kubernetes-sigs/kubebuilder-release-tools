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

// RevListOptions allow to provide the options for a `git rev-list ...` command.
// NOTE: only an incomplete set of options are implemented
type RevListOptions struct {
	Committish Committish

	Merges     bool
	NoMerges   bool
	MinParents *int
	MaxParents *int

	Pretty string
}

// validate implements cliOptions.validate.
func (opts RevListOptions) validate() error {
	if opts.Committish == nil {
		return fmt.Errorf("a committish must be provided")
	}

	if opts.Merges && opts.MinParents != nil {
		return fmt.Errorf("--merges (equivalent to --min-parent=2) and --min-parents provided")
	}
	if opts.NoMerges && opts.MaxParents != nil {
		return fmt.Errorf("--no-merges (equivalent to --max-parent=1) and --max-parents provided")
	}

	return nil
}

// arguments implements cliOptions.arguments.
func (opts RevListOptions) arguments() (args []string) {
	args = append(args, "rev-list")

	args = append(args, opts.Committish.Committish())

	if opts.Merges {
		args = append(args, "--merges")
	} else if opts.MinParents != nil {
		args = append(args, fmt.Sprintf("--min-parents=%d", *opts.MinParents))
	}
	if opts.NoMerges {
		args = append(args, "--no-merges")
	} else if opts.MaxParents != nil {
		args = append(args, fmt.Sprintf("--max-parents=%d", *opts.MaxParents))
	}

	if opts.Pretty != "" {
		args = append(args, fmt.Sprintf("--pretty=%s", opts.Pretty))
	}

	return
}
