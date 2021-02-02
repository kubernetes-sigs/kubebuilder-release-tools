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

// RemoteGetUrlOptions allow to provide the options for a `git remote get-url ...` command.
type RemoteGetUrlOptions struct {
	Remote string

	Push bool
	All  bool
}

// validate implements cliOptions.validate.
func (opts RemoteGetUrlOptions) validate() error {
	if opts.Remote == "" {
		return fmt.Errorf("a remote must be provided")
	}

	return nil
}

// arguments implements cliOptions.arguments.
func (opts RemoteGetUrlOptions) arguments() (args []string) {
	args = append(args, "remote")
	args = append(args, "get-url")

	args = append(args, opts.Remote)

	if opts.All {
		args = append(args, "--all")
	} else if opts.Push {
		args = append(args, "--push")
	}

	return
}
