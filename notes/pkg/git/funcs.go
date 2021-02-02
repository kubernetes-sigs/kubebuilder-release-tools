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

/* Direct CLI calls */

// Describe executes `git describe ...` with the provided options.
func Describe(opts DescribeOptions) (string, error) {
	return Command.Describe(opts)
}

// Fetch executes `git fetch ...` with the provided options.
func Fetch(opts FetchOptions) (string, error) {
	return Command.Fetch(opts)
}

// ForEachRef executes `git for-each-ref ...` with the provided options.
func ForEachRef(opts ForEachRefOptions) (string, error) {
	return Command.ForEachRef(opts)
}

// RemoteGetUrl executes `git remote get-url ...` with the provided options.
func RemoteGetUrl(opts RemoteGetUrlOptions) (string, error) {
	return Command.RemoteGetUrl(opts)
}

// RevList executes `git rev-list ...` with the provided options.
func RevList(opts RevListOptions) (string, error) {
	return Command.RevList(opts)
}

// RevParse executes `git rev-parse ...` with the provided options.
func RevParse(opts RevParseOptions) (string, error) {
	return Command.RevParse(opts)
}
