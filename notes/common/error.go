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
	"errors"
	"fmt"
	"os/exec"
)

// ErrOut wraps exec.ExitErrors so that the message displays their
// stderr output.  If the error is not an exist error, or does not
// wrap one, this returns the error without any changes.
func ErrOut(err error) error {
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

func (e errOut) Error() string {
	return fmt.Sprintf("[%v] %q", e.actual.Error(), string(e.actual.Stderr))
}

func (e errOut) Unwrap() error {
	return e.actual
}
