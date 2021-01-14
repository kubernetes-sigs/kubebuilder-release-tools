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

// RevParseOptions allow to provide the options for a `git rev-parse ...` command.
// NOTE: only an incomplete set of options are implemented
type RevParseOptions struct {
	Committish Committish

	AbbrevRef        bool
	Symbolic         bool
	SymbolicFullName bool
}

// validate implements cliOptions.validate.
func (opts RevParseOptions) validate() error {
	if opts.Committish == nil {
		return fmt.Errorf("a committish must be provided")
	}

	if opts.Symbolic && opts.SymbolicFullName {
		return fmt.Errorf("symbolic and symbolic full name options provided")
	}

	return nil
}

// arguments implements cliOptions.arguments.
func (opts RevParseOptions) arguments() (args []string) {
	args = append(args, "rev-parse")

	if opts.AbbrevRef {
		args = append(args, "--abrev-ref")
	} else if opts.Symbolic {
		args = append(args, "--symbolic")
	} else if opts.SymbolicFullName {
		args = append(args, "--symbolic-full-name")
	}

	args = append(args, opts.Committish.Committish())

	return
}
