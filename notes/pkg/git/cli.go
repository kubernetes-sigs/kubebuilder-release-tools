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
	"errors"
	"fmt"
	"os/exec"
)

// CLI provides the git CLI interface allowing it to be mocked for tests.
type CLI interface {
	// Describe executes `git describe ...` commands.
	Describe(DescribeOptions) (string, error)
	// Fetch executes `git fetch ...` commands.
	Fetch(FetchOptions) (string, error)
	// ForEachref executes `git for-each-ref ...` commands.
	ForEachRef(ForEachRefOptions) (string, error)
	// RemoteGetUrl executes `git remote get-url ...` commands.
	RemoteGetUrl(RemoteGetUrlOptions) (string, error)
	// RevList executes `git rev-list ...` commands.
	RevList(RevListOptions) (string, error)
	// RevParse executes `git rev-parse ...` commands.
	RevParse(RevParseOptions) (string, error)
}

// wrapExistErrors wraps exec.ExitErrors so that the message displays their
// stderr output. If the error is not an exist error, or does not wrap one,
// this returns the error without any changes.
func wrapExistErrors(err error) error {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return err
	}
	return errOut{actual: exitErr}
}

// errOut is an Error that prints the underlying ExitError's stderr in addition
// to the normal message.
type errOut struct {
	actual *exec.ExitError
}

// Error implements error.Error.
func (e errOut) Error() string {
	return fmt.Sprintf("[%v] %q", e.actual.Error(), string(e.actual.Stderr))
}

// Unwrap implements errors.Wrapper.Unwrap.
func (e errOut) Unwrap() error {
	return e.actual
}

type cliOptions interface {
	validate() error
	arguments() []string
}

func execute(opts cliOptions) (string, error) {
	if err := opts.validate(); err != nil {
		return "", err
	}

	cmd := exec.Command("git", opts.arguments()...)
	b, err := cmd.Output()
	return string(b), wrapExistErrors(err)
}

var Command CLI = cli{}

type cli struct{}

// Describe implements CLI.Describe.
func (cli) Describe(opts DescribeOptions) (string, error) {
	return execute(opts)
}

// Fetch implements CLI.Fetch.
func (cli) Fetch(opts FetchOptions) (string, error) {
	return execute(opts)
}

// ForEachRef implements CLI.ForEachRef.
func (cli) ForEachRef(opts ForEachRefOptions) (string, error) {
	return execute(opts)
}

// RemoteGetUrl implements CLI.RemoteGetUrl.
func (cli) RemoteGetUrl(opts RemoteGetUrlOptions) (string, error) {
	return execute(opts)
}

// RevList implements CLI.RevList.
func (cli) RevList(opts RevListOptions) (string, error) {
	return execute(opts)
}

// RevParse implements CLI.RevParse.
func (cli) RevParse(opts RevParseOptions) (string, error) {
	return execute(opts)
}
