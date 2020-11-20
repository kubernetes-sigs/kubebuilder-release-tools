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

type Logger interface {
	Debug(content string)
	Debugf(format string, args ...interface{})
	Info(content string)
	Infof(format string, args ...interface{})
	Warning(content string)
	Warningf(format string, args ...interface{})
	Error(content string)
	Errorf(format string, args ...interface{})
	Fatal(exitCode int, content string)
	Fatalf(exitCode int, format string, args ...interface{})
}
