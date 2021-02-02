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

// Ensure that CLIMock implements CLI at compile time.
var _ CLI = CLIMock{}

// CLIMock implements CLI using the functions provided in the fields instead of the actual cli.
// It is meant to allow unit tests of packages that make use of the CLI interface.
type CLIMock struct {
	DescribeF     func(DescribeOptions) (string, error)
	FetchF        func(options FetchOptions) (string, error)
	ForEachRefF   func(options ForEachRefOptions) (string, error)
	RemoteGetUrlF func(options RemoteGetUrlOptions) (string, error)
	RevListF      func(RevListOptions) (string, error)
	RevParseF     func(RevParseOptions) (string, error)
}

// CLIMock.Describe implements CLI.Describe
func (mock CLIMock) Describe(opts DescribeOptions) (string, error) {
	if mock.DescribeF == nil {
		panic("Describe not expected")
	}
	return mock.DescribeF(opts)
}

// CLIMock.Fetch implements CLI.Fetch
func (mock CLIMock) Fetch(opts FetchOptions) (string, error) {
	if mock.FetchF == nil {
		panic("Fetch not expected")
	}
	return mock.FetchF(opts)
}

// CLIMock.ForEachRef implements CLI.ForEachRef
func (mock CLIMock) ForEachRef(opts ForEachRefOptions) (string, error) {
	if mock.ForEachRefF == nil {
		panic("ForEachRef not expected")
	}
	return mock.ForEachRefF(opts)
}

// CLIMock.RemoteGetUrl implements CLI.RemoteGetUrl.
func (mock CLIMock) RemoteGetUrl(opts RemoteGetUrlOptions) (string, error) {
	if mock.RemoteGetUrlF == nil {
		panic("RemoteGetUrl not expected")
	}
	return mock.RemoteGetUrlF(opts)
}

// CLIMock.RevList implements CLI.RevList
func (mock CLIMock) RevList(opts RevListOptions) (string, error) {
	if mock.RevListF == nil {
		panic("RevList not expected")
	}
	return mock.RevListF(opts)
}

// CLIMock.RevParse implements CLI.RevParse
func (mock CLIMock) RevParse(opts RevParseOptions) (string, error) {
	if mock.RevParseF == nil {
		panic("RevParse not expected")
	}
	return mock.RevParseF(opts)
}
