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

/*
This package provides a pure Go API with part of the `git` command.

Each implemented subcommand is represented as a method of the interface CLI in cli.go.
This interface is implemented by Command and also exported as package level funtions.

Example:
	package main

	import (
		"fmt"

		"sigs.k8s.io/kubebuilder-release-tools/notes/git"
	)

	func describeExportedVariable() {
		output, err := git.Command.Describe(git.DescribeOptions{
			Committish: Head,
			Tags:       true, // --tags
		})
		if err != nil {
			fmt.Println(fmt.Errorf("unable to call `git describe --tags HEAD`: %w", err))
		} else {
			fmt.Println(output)
		}
	}

	func describeExportedFunctions() {
		output, err := git.Describe(git.DescribeOptions{
			Committish: Head,
			Tags:       true, // --tags
		})
		if err != nil {
			fmt.Println(fmt.Errorf("unable to call `git describe --tags HEAD`: %w", err))
		} else {
			fmt.Println(output)
		}
	}

	func main() {
		describeExportedVariable()
		describeExportedFunctions()
	}

A higher level interface is also povided by the Utilities interface in utils.go.
This interface is implemented by the Utils exported variable.

Both CLI and Utilities interfaces also have mock implementations (CLIMock and UtilitiesMock
respectively) that allow to use user provided functions in the object fields instead of actual
command calls. These mocks are intended for testing purposes.
*/
package git
