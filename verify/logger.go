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

package verify

import (
	"fmt"
	"strings"
)

const (
	errorPrefix   = "::error::"
	debugPrefix   = "::debug::"
	warningPrefix = "::debug::"
)

type logger struct{}

func (logger) log(prefix, content string) {
	for _, s := range strings.Split(content, "\n") {
		fmt.Println(prefix + s)
	}
}

func (l logger) errorf(format string, args ...interface{}) {
	l.log(errorPrefix, fmt.Sprintf(format, args...))
}

func (l logger) debugf(format string, args ...interface{}) {
	l.log(debugPrefix, fmt.Sprintf(format, args...))
}

func (l logger) warningf(format string, args ...interface{}) {
	l.log(warningPrefix, fmt.Sprintf(format, args...))
}
