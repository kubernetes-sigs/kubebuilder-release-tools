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

// ForEachRefOptions allow to provide the options for a `git for-each-ref ...` command.
// NOTE: only an incomplete set of options are implemented
type ForEachRefOptions struct {
	Committish Committish

	Format string
}

// validate implements cliOptions.validate.
func (opts ForEachRefOptions) validate() error {
	if opts.Committish == nil {
		return fmt.Errorf("a committish must be provided")
	}

	return nil
}

// arguments implements cliOptions.arguments.
func (opts ForEachRefOptions) arguments() (args []string) {
	args = append(args, "for-each-ref")

	if opts.Format != "" {
		args = append(args, fmt.Sprintf("--format=%q", opts.Format))
	}

	args = append(args, opts.Committish.Committish())

	return
}
