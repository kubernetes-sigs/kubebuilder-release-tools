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

// FetchOptions allow to provide the options for a `git fetch ...` command.
// NOTE: only an incomplete set of options are implemented.
type FetchOptions struct {
	Remote string

	Tags bool
}

// validate implements cliOptions.validate.
func (opts FetchOptions) validate() error {
	return nil
}

// arguments implements cliOptions.arguments.
func (opts FetchOptions) arguments() (args []string) {
	args = append(args, "fetch")

	if opts.Remote != "" {
		args = append(args, opts.Remote)
	}

	if opts.Tags {
		args = append(args, "--tags")
	}

	return
}
