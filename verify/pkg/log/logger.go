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

package log

import (
	"fmt"
	"os"
	"strings"
)

// Prefixes used in GitHub actions
const (
	ghPrefixDebug   = "::debug::"
	ghPrefixWarning = "::warning::"
	ghPrefixError   = "::error::"
)

// Verify that logger implements Logger
var _ Logger = logger{}

// logger provides logging functions for GitHub actions
type logger struct{
	// Prefixes for the different logging methods
	prefixDebug   string
	prefixInfo    string
	prefixWarning string
	prefixError   string
}

// New returns a basic logger for GitHub actions
func New() Logger {
	return &logger{
		prefixDebug:   ghPrefixDebug,
		prefixInfo:    "",
		prefixWarning: ghPrefixWarning,
		prefixError:   ghPrefixError,
	}
}

// NewFor returns a named logger for GitHub actions
func NewFor(name string) Logger {
	return &logger{
		prefixDebug:   fmt.Sprintf("%s[%s]", ghPrefixDebug, name),
		prefixInfo:    fmt.Sprintf("[%s]", name),
		prefixWarning: fmt.Sprintf("%s[%s]", ghPrefixWarning, name),
		prefixError:   fmt.Sprintf("%s[%s]", ghPrefixError, name),
	}
}

func (l logger) prefixFor(level LoggingLevel) string {
	switch level {
	case Debug:
		return l.prefixDebug
	case Info:
		return l.prefixInfo
	case Warning:
		return l.prefixWarning
	case Error:
		return l.prefixError
	default:
		panic("invalid logging level")
	}
}

func (l logger) log(content string, level LoggingLevel) {
	prefix := l.prefixFor(level)
	for _, s := range strings.Split(content, "\n") {
		fmt.Println(prefix + s)
	}
}

func (l logger) Debug(content string) {
	l.log(content, Debug)
}

func (l logger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

func (l logger) Info(content string) {
	l.log(content, Info)
}

func (l logger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l logger) Warning(content string) {
	l.log(content, Warning)
}

func (l logger) Warningf(format string, args ...interface{}) {
	l.Warning(fmt.Sprintf(format, args...))
}

func (l logger) Error(content string) {
	l.log(content, Error)
}

func (l logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l logger) Fatal(exitCode int, content string) {
	l.log(content, Error)
	os.Exit(exitCode)
}

func (l logger) Fatalf(exitCode int, format string, args ...interface{}) {
	l.Fatal(exitCode, fmt.Sprintf(format, args...))
}
