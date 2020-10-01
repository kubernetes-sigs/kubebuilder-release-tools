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

/*
This tool prints all the titles of all PRs from previous release to HEAD.
This needs to be run *before* a tag is created.

Use these as the base of your release notes.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"sigs.k8s.io/kubebuilder-release-tools/notes/compose"
	"sigs.k8s.io/kubebuilder-release-tools/notes/git"
)

var (
	fromTag          = flag.String("from", "", "The tag or commit to start from.")
	branchName       = flag.String("branch", "", "The release branch to run on (defaults to current)")
	showOthers       = flag.String("show-others", "", "Comma-separate set of non-code changes to show (docs,infra)")
	project          = flag.String("project", "", "GitHub project in org/repo form to use to generate link to past releases (defaults to a value extracted from the remote of the branch or 'upstream'")
	useUpstreams     = flag.Bool("use-upstream", true, "try to compose information from upstream versions of the local release branches")
	refreshUpstreams = flag.Bool("refresh-upstream", true, "git-fetch the remote for the current branch before continuing (only relevant if use-upstream is set)")
	relType          = flag.String("r", "final", "type of release -- final, alpha, beta, or rc")
	forceV1          = flag.Bool("force-v1", false, "if the current release is 0.Y-style, assume the next 'major' release is 1.0 instead of being 0.Y-style")
	extraInfoOnFinal = flag.Bool("print-full-final", true, "if the current release would bring us from pre-release to final, print the full changes since the last final release")
)

// run wraps what would otherwise be main to have one error handler with
// detailed stderr on exec errors
func run() error {
	if *fromTag == "" {
		var err error
		*branchName, err = git.Actual.CurrentBranch()
		if err != nil {
			return err
		}
	}
	log.Printf("starting from branch %q", *branchName)

	branch, err := compose.ReleaseFromBranch(*branchName)
	if err != nil {
		return err
	}

	if *useUpstreams {
		branch.UseUpstream = true
		if *refreshUpstreams {
			if err := refreshUpstream(*branchName); err != nil {
				// this might happen if we're on a new branch, so don't fret
				fmt.Fprintf(os.Stderr, "\x1b[1;31munable to refresh upstream, continuing on without it -- you may want to do this manually\x1b[0m: %v\n", err)
			}
		}
	}

	var (
		changes compose.ChangeLog
		since   git.Committish
	)
	if *fromTag == "" {
		changes, since, err = compose.Changes(git.Actual, &branch)
	} else {
		since = git.SomeCommittish(*fromTag)
		changes, err = compose.ChangesSince(git.Actual, branch, since)
	}
	if err != nil {
		return err
	}

	if *project == "" {
		var err error
		if branch.UseUpstream {
			// reset UseUpstream so we don't try to get the remote for an upstream itself
			*project, err = findProject(compose.ReleaseBranch{Version: branch.Version}.String())
		} else {
			log.Printf("current branch %q has no assicated upstream, assuming upstream remote is \"upstream\" for auto-setting project", branch)
			*project, err = findProject("")
		}
		if err != nil {
			log.Printf("unable to determine URL for upstream remote (set --project manually): %v", err)
		}
	}

	return printLog(branch, logChunk{ChangeLog: changes, since: since})
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `Usage of %[1]s [FLAGS]:

  Examples:

  # Prep for a beta release 
  %[1]s -r beta

  # Prep for a release that bumps version 0.Y to 1.0.0
  %[1]s --force-v1

  # Show docs contributions in the release notes
  %[1]s --show-others docs

  Flags:

`, os.Args[0])

		flag.PrintDefaults()
	}
	flag.Parse()

	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

// logChunk is a piece of a full commit log.  It contains one set of changes
// since a given committish.
type logChunk struct {
	since git.Committish
	compose.ChangeLog
}

// Print prints the changes within this chunk along with a header indicating
// when these changes are from.
func (c *logChunk) Print() {
	fmt.Printf("\n**changes since [%[1]s](https://github.com/%[2]s/releases/%[1]s)**\n", c.since.Committish(), *project)

	sectionIfPresent(c.Breaking, ":warning: Breaking Changes")
	sectionIfPresent(c.Features, ":sparkles: New Features")
	sectionIfPresent(c.Bugs, ":bug: Bug Fixes")

	optionals := strings.Split(*showOthers, ",")
	for _, opt := range optionals {
		switch opt {
		case "docs":
			sectionIfPresent(c.Docs, ":book: Documentation")
		case "infra":
			sectionIfPresent(c.Infra, ":seedling: Infra & Such")
		case "":
			// don't do anything
		default:
			log.Printf("unknown optinal section %q, skipping", opt)
		}
	}

	sectionIfPresent(c.Uncategorized, ":question: Sort these by hand")
}

// release holds the name of the upcoming release, and the intermediate information
// used to make that decision.
type release struct {
	compose.ReleaseInfo
	next compose.ReleaseTag
}

// releaseInfo computes compose.ReleaseInfo & the expected next release version
// given a branch and some changes.
func releaseInfo(branch compose.ReleaseBranch, changes logChunk) (release, error) {
	relInfo := compose.ReleaseInfo{Pre10: !*forceV1}
	switch *relType {
	case "final":
		relInfo.Kind = compose.ReleaseFinal
	case "alpha":
		relInfo.Kind = compose.ReleaseAlpha
	case "beta":
		relInfo.Kind = compose.ReleaseBeta
	case "rc":
		relInfo.Kind = compose.ReleaseCandidate
	default:
		return release{}, fmt.Errorf("unknown release type %q, must be final|alpha|beta|rc", *relType)
	}
	nextVer, err := changes.ExpectedNextVersion(changes.since, relInfo)
	if err != nil {
		return release{}, err
	}

	return release{
		ReleaseInfo: relInfo,
		next:        nextVer,
	}, nil
}

// printLog prints the release log with appropriate header, changes-since link(s),
// and potentially a full extra change-log if we're going from pre-release to final.
func printLog(branch compose.ReleaseBranch, recentChanges logChunk) error {
	if len(recentChanges.Breaking) > 0 {
		fmt.Fprint(os.Stderr, "\x1b[1;31mbreaking changes this version\x1b[0m\n")
	}
	if len(recentChanges.Uncategorized) > 0 {
		fmt.Fprint(os.Stderr, "\x1b[1;35munknown changes in this release -- categorize manually\x1b[0m\n")
	}

	rel, err := releaseInfo(branch, recentChanges)
	if err != nil {
		return err
	}

	// if we're going from pre-release to final, print out the total changes
	var otherChanges *logChunk
	if *extraInfoOnFinal && compose.IsPreReleaseToFinal(recentChanges.since, rel.next) {
		// the cast is guaranteed by IsPreReleaseFinal
		prev, err := compose.ClosestFinal(git.Actual, recentChanges.since.(compose.ReleaseTag))
		if err != nil {
			return fmt.Errorf("unable to find last final release (try running with --print-full-final=false if that's expected): %w", err)
		}

		otherLog, err := compose.ChangesSince(git.Actual, branch, *prev)
		if err != nil {
			return fmt.Errorf("unable to compute changes since last final release (try running with --print-full-final=false if that's expected): %w", err)
		}
		otherChanges = &logChunk{
			ChangeLog: otherLog,
			since:     *prev,
		}
	}

	// the actual log
	fmt.Printf("# %s\n", rel.next)

	recentChanges.Print()

	if otherChanges != nil {
		otherChanges.Print()
	}

	fmt.Println("")
	fmt.Println("*Thanks to all our contributors!*")

	return nil
}

// formatEntry turns out a single log entry into a string form for printing.
func formatEntry(entry compose.LogEntry) string {
	if entry.PRNumber == "" {
		return entry.Title
	}
	return fmt.Sprintf("%s (#%s)", entry.Title, entry.PRNumber)
}

// sectionIfPresent prints a section with the given title if any changes are
// present.
func sectionIfPresent(changes []compose.LogEntry, title string) {
	if len(changes) > 0 {
		fmt.Println("")
		fmt.Printf("## %s\n", title)
		fmt.Println("")
		for _, change := range changes {
			fmt.Printf("- %s\n", formatEntry(change))
		}
	}
}

// findProject guesses at the project for this repo. If a branch name is
// specified, it will be extracted from a github remote on the remote for the
// upstream for that branch.  Otherwise, it'll be extracted from a github
// remote on the "upstream" remote.
func findProject(branchName string) (string, error) {
	remote := "upstream"
	if branchName != "" {
		var err error
		remote, err = git.Actual.RemoteForUpstreamFor(branchName)
		if err != nil {
			return "", fmt.Errorf("unable to determine upstream of branch %q: %w", branchName, err)
		}
		log.Printf("remote for branch %q was %q", branchName, remote)
	}

	log.Printf("checking upstream URL for remote %q", remote)
	upstreamURL, err := git.Actual.URLForRemote(remote)
	if err != nil {
		return "", fmt.Errorf("unable to determine upstream URL for %q: %w", remote, err)
	}

	project := upstreamURL

	project = strings.TrimPrefix(project, "git@github.com:")
	project = strings.TrimPrefix(project, "https://github.com/")
	if project == upstreamURL {
		return "", fmt.Errorf("unrecognized upstream URL format %q (expected either git@github.com:* or https://github.com/*)", upstreamURL)
	}

	project = strings.TrimSuffix(project, ".git")
	return project, nil
}

func refreshUpstream(branchName string) error {
	remote, err := git.Actual.RemoteForUpstreamFor(branchName)
	if err != nil {
		fmt.Errorf("unable to determine upstream of branch %q: %w", branchName, err)
	}
	if err := git.Actual.Fetch(remote); err != nil {
		return fmt.Errorf("unable to refresh remote %q: %w", remote, err)
	}
	return nil
}
