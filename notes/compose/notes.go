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

package compose

import (
	"fmt"
	golog "log"
	"regexp"
	"strconv"
	"strings"

	"github.com/blang/semver/v4"

	"sigs.k8s.io/kubebuilder-release-tools/notes/common"
	"sigs.k8s.io/kubebuilder-release-tools/notes/git"
)

var (
	releaseRE = regexp.MustCompile(`^release-((?:0\.(?P<minor>[[:digit:]]+))|(?P<major>[[:digit:]]+))$`)
)

// TODO(directxman12): we could use go-git, but it doesn't implement
// git-describe, which is a pain to implement by hand.

// ReleaseFromBranch extracts a major-ish (X or 0.Y) release given a branch name.
func ReleaseFromBranch(branchName string) (ReleaseBranch, error) {
	parts := releaseRE.FindStringSubmatch(branchName)
	if parts == nil {
		return ReleaseBranch{}, fmt.Errorf("%q is not a valid release branch (release-0.Y or release-X)", branchName)
	}
	minorRaw := parts[releaseRE.SubexpIndex("minor")]
	majorRaw := parts[releaseRE.SubexpIndex("major")]
	switch {
	case minorRaw != "":
		minor, err := strconv.ParseUint(minorRaw, 10, 64)
		if err != nil {
			return ReleaseBranch{}, fmt.Errorf("could not parse minor version from %q: %w", minorRaw, err)
		}
		if minor == 0 {
			return ReleaseBranch{}, fmt.Errorf("release-0.0 is not a valid release")
		}
		return ReleaseBranch{
			Version: semver.Version{Major: 0, Minor: minor},
		}, nil
	case majorRaw != "":
		major, err := strconv.ParseUint(majorRaw, 10, 64)
		if err != nil {
			return ReleaseBranch{}, fmt.Errorf("could not parse major version from %q: %w", majorRaw, err)
		}
		if major == 0 {
			return ReleaseBranch{}, fmt.Errorf("release-0 is not a valid release")
		}
		return ReleaseBranch{
			Version: semver.Version{Major: major},
		}, nil
	default:
		return ReleaseBranch{}, fmt.Errorf("%q is not a valid release branch (release-0.Y or release-X)", branchName)
	}
}

// ReleaseBranch represents a branch associated with major-ish (X or 0.Y) set
// of releases.
type ReleaseBranch struct {
	semver.Version
	UseUpstream bool
}

func (b ReleaseBranch) String() string {
	upstreamPart := ""
	if b.UseUpstream {
		upstreamPart = "@{u}"
	}
	if b.Major == 0 {
		return fmt.Sprintf("release-0.%d%s", b.Minor, upstreamPart)
	}
	return fmt.Sprintf("release-%d%s", b.Major, upstreamPart)
}
func (b ReleaseBranch) Committish() string {
	return b.String()
}

// ReleaseTag is a Committish that's actually a version-tag for a release.
type ReleaseTag semver.Version

func (v ReleaseTag) Committish() string {
	return "v" + semver.Version(v).String()
}
func (v ReleaseTag) String() string {
	return v.Committish()
}
func (v ReleaseTag) Validate() error {
	if len(v.Pre) != 0 && len(v.Pre) != 2 {
		return fmt.Errorf("invalid pre-release info (must be -{alpha,beta,rc}.version)")
	}
	if len(v.Pre) == 2 && (v.Pre[0].IsNum || !v.Pre[1].IsNum) {
		// TODO: validate alpha, beta, rc
		return fmt.Errorf("invalid pre-release info (must be -{alpha,beta,rc}.version)")
	}
	return nil
}

// FirstCommit is a Committish that's the first commit on a branch, generally
// used when the previous release tag does not exist.
type FirstCommit struct {
	Commit git.Commit
	Branch ReleaseBranch
}

func (c FirstCommit) Committish() string {
	return c.Commit.Committish()
}

// parseReleaseTag parses a git tag name into a ReleaseTag.
func parseReleaseTag(tagRaw git.Tag) (*ReleaseTag, error) {
	tagRawBytes := []byte(tagRaw)
	if tagRawBytes[0] != 'v' {
		return nil, fmt.Errorf("not a version tag (vX.Y.Z)")
	}
	tagRawBytes = tagRawBytes[1:] // skip the 'v'

	tagParsed, err := semver.Parse(string(tagRawBytes))
	if err != nil {
		return nil, err
	}
	tag := ReleaseTag(tagParsed)
	if err := tag.Validate(); err != nil {
		return nil, err
	}
	return &tag, nil
}

// LatestRelease returns the most recent ReleaseTag on this branch, or a the
// FirstCommit if none existed.
func (b ReleaseBranch) LatestRelease(gitImpl git.Git, checkVersion bool) (git.Committish, error) {
	tagRaw, err := gitImpl.ClosestTag(b)
	if err != nil {
		golog.Printf("unable to get latest tag starting at %q, assuming we need to look for the first commit instead (%v)", b, err)
		// try to get the first commit
		commitSHA, commitErr := gitImpl.FirstCommit(b.String())
		if commitErr != nil {
			// double wrap to get both errors
			return nil, fmt.Errorf("unable to grab first commit on branch %q (%v), also unable to fetch most recent tag: %w", b, err, commitErr)
		}
		return FirstCommit{
			Branch: b,
			Commit: commitSHA,
		}, nil
	}

	tag, err := parseReleaseTag(tagRaw)
	if err != nil {
		return nil, fmt.Errorf("latest tag %q on branch %q was not a (valid?) version: %w", tag, b, err)
	}

	if !checkVersion {
		golog.Printf("latest release on branch %q is probably %q", b, tag)
		return ReleaseTag(*tag), nil
	}

	golog.Printf("latest release on branch %q is probably %q", b, tag)
	relTag := ReleaseTag(*tag)
	return relTag, b.VerifyTagBelongs(relTag)
}

// VerifyTagBelongs checks that a given tag has the correct major-ish version
// for this branch.
func (b ReleaseBranch) VerifyTagBelongs(tag ReleaseTag) error {
	if tag.Major != b.Major || (tag.Major == 0 && tag.Minor != b.Minor) {
		return fmt.Errorf("tag's version %v does not match the branch's version %v", tag, b)
	}
	return nil
}

// checkOrClearUpstream verifies that the upstream exists for this branch and
// clears UseUpstream if it does not.  If UseUpstream is already false, this is
// a no-op.
func checkOrClearUpstream(gitImpl git.Git, branch *ReleaseBranch) {
	if !branch.UseUpstream {
		return
	}
	if err := gitImpl.HasUpstream(branch.String()); err != nil {
		branch.UseUpstream = false
		golog.Printf("branch %q did not have an upstream, falling back to non-upstream (%v)", branch, err)
	}
}

// CurrentVersion locates the closest current version (release tag or first
// commit), starting at the HEAD of the current branch.  If the branch has an
// upstream and is configured to use it, it'll try that first.  If that doesn't
// work, it'll clear the UseUpstream field and try the non-upstream version.
//
// Furthermore, if it looks like the closest release actually shoulbelongs to the previous
// release branch, it'll double-check that branch instead, to get the actual most recent release.
// For instance, on a fresh `release-0.7` branch, the "latest" release might be `v0.6.0`
// (since the `v0.Y.0` releases are always off of the main branch), so it'll check `release-0.6`
// to find that the *actual* latest release is `v0.6.3`.
func CurrentVersion(gitImpl git.Git, branch *ReleaseBranch) (git.Committish, error) {
	origUseUpstream := branch.UseUpstream // keep this around to keep trying later if necessary
	checkOrClearUpstream(gitImpl, branch)

	latestHere, err := branch.LatestRelease(gitImpl, false)
	if err != nil {
		return nil, err
	}

	tag, isTag := latestHere.(ReleaseTag)
	if !isTag {
		golog.Printf("no latest tag, not double-checking version matches")
		return latestHere, nil
	}

	switch {
	case branch.Major == 0 && tag.Major == 0 && tag.Minor == branch.Minor-1:
		// most recent tag is a release behind, check the previous branch:
		// on the first release on a release branch, we'll generally end up
		// seeing the first release of the last "major-ish" (X, or 0.Y), since
		// that'll be the only one that ends up on master (the rest are on
		// a release branch).  Therefore, switch branches backwards to get the actual
		// last tag.
		prevRel := ReleaseBranch{
			Version:     semver.Version{Major: 0, Minor: tag.Minor},
			UseUpstream: origUseUpstream,
		}
		golog.Printf("most recent tag %q is from last version (probably a 0.Y bump), double-checking previous release branch %q for actual latest version", tag.Committish(), prevRel)
		checkOrClearUpstream(gitImpl, &prevRel)
		return prevRel.LatestRelease(gitImpl, true)
	case branch.Major == 1 && tag.Major == 0:
		// ditto as above, except with 1 releases instead to 0.Y
		prevRel := ReleaseBranch{
			Version:     semver.Version{Minor: tag.Minor},
			UseUpstream: branch.UseUpstream,
		}
		golog.Printf("most recent tag %q is from last version (probably a 0.Y --> 1 bump), double-checking previous release branch %q for actual latest version", tag.Committish(), prevRel)
		checkOrClearUpstream(gitImpl, &prevRel)
		return prevRel.LatestRelease(gitImpl, true)
	case branch.Major > 0 && tag.Major == branch.Major-1:
		// same as the first case, except with X releases instead of 0.Y
		prevRel := ReleaseBranch{
			Version:     semver.Version{Major: tag.Major},
			UseUpstream: branch.UseUpstream,
		}
		golog.Printf("most recent tag %q is from last version (probably a X bump), double-checking previous release branch %q for actual latest version", tag.Committish(), prevRel)
		checkOrClearUpstream(gitImpl, &prevRel)
		return prevRel.LatestRelease(gitImpl, true)
	default:
		return tag, branch.VerifyTagBelongs(tag)
	}
}

// LogEntry contains a single changelog entry from a PR.
type LogEntry struct {
	PRNumber string
	Title    string
}

// ChangeLog holds all changes between a release and HEAD, organized by release type.
type ChangeLog struct {
	Breaking      []LogEntry
	Features      []LogEntry
	Bugs          []LogEntry
	Docs          []LogEntry
	Infra         []LogEntry
	Uncategorized []LogEntry
}

// entryFromCommit adds a changelog entry to this changelog
// based on the emoji marker in the title.
func (l *ChangeLog) entryFromCommit(prNum, title string) {
	entry := LogEntry{PRNumber: prNum}

	prType, title := common.PRTypeFromTitle(title)
	entry.Title = title
	switch prType {
	case common.FeaturePR:
		l.Features = append(l.Features, entry)
	case common.BugfixPR:
		l.Bugs = append(l.Bugs, entry)
	case common.DocsPR:
		l.Docs = append(l.Docs, entry)
	case common.InfraPR:
		l.Infra = append(l.Infra, entry)
	case common.BreakingPR:
		l.Breaking = append(l.Breaking, entry)
	case common.UncategorizedPR:
		l.Uncategorized = append(l.Uncategorized, entry)
	default:
		panic(fmt.Sprintf("unrecognized internal PR type %v", prType))
	}
}

// ReleaseKind indicates the "finality" of this release -- pre-release (alpha,
// beta, rc) or final.
type ReleaseKind int

const (
	ReleaseFinal     ReleaseKind = 0
	ReleaseAlpha     ReleaseKind = 1
	ReleaseBeta      ReleaseKind = 2
	ReleaseCandidate ReleaseKind = 3
)

// ReleaseInfo describes the desired type of release.
type ReleaseInfo struct {
	// Kind is the finality of the release.
	Kind ReleaseKind
	// Pre10 indicates that if the current release is 0.Y, and we'd need a new
	// major-ish version, choose v0.(Y+1) and not v1.0.0.
	Pre10 bool
}

// ExpectedNextVersion computes what the next version for should be given a set
// of changes, and desired type of release.
//
// Roughly, this means that, if one of the releases (current or desired next)
// is a final release:
//
// - 0.Y releases are equivalent to either X or X.Y releases
// - Breaking changes bump X
// - Features bump Y
// - Anything else just bumps Z
//
// If we're jumping between prereleases, ignore all that and either increment
// the pre-release number or reset the number to zero if we're switching types.
//
// If Pre10 is set, never jump to v1.0.0.
func (c ChangeLog) ExpectedNextVersion(currentVersion git.Committish, info ReleaseInfo) (ReleaseTag, error) {
	tag, isTag := currentVersion.(ReleaseTag)
	if !isTag {
		res := ReleaseTag(semver.Version{
			Minor: 1,
		})
		switch info.Kind {
		case ReleaseAlpha:
			res.Pre = []semver.PRVersion{{VersionStr: "alpha"}, {VersionNum: 0, IsNum: true}}
		case ReleaseBeta:
			res.Pre = []semver.PRVersion{{VersionStr: "beta"}, {VersionNum: 0, IsNum: true}}
		case ReleaseCandidate:
			res.Pre = []semver.PRVersion{{VersionStr: "rc"}, {VersionNum: 0, IsNum: true}}
		}
		return res, nil
	}

	// final releases
	newTag := tag
	if info.Kind == ReleaseFinal {
		// pre --> final: reset pre, keep version
		if len(newTag.Pre) > 0 {
			newTag.Pre = nil
			return newTag, nil
		}

		// final --> final: bump according to rules
		return c.nextFinalVersion(tag, info.Pre10), nil
	}

	// easy pre-release case: same type of pre-release
	// alpha --> alpha || beta --> beta || rc --> rc
	wasPre := len(tag.Pre) > 0
	alphaToAlpha := wasPre && tag.Pre[0] == semver.PRVersion{VersionStr: "alpha"} && info.Kind == ReleaseAlpha
	betaToBeta := wasPre && tag.Pre[0] == semver.PRVersion{VersionStr: "beta"} && info.Kind == ReleaseBeta
	candidateToCandidate := wasPre && tag.Pre[0] == semver.PRVersion{VersionStr: "candidate"} && info.Kind == ReleaseCandidate
	if alphaToAlpha || betaToBeta || candidateToCandidate {
		newTag := tag
		// don't clobber old release
		newTag.Pre = make([]semver.PRVersion, len(tag.Pre))
		copy(newTag.Pre, tag.Pre)
		newTag.Pre[1].VersionNum++
		return newTag, nil
	}

	// otherwise, if the old release was a final release...
	if tag.Pre == nil {
		// ...bump according to rules...
		newTag = c.nextFinalVersion(tag, info.Pre10)
	}

	// ...either way, add the appropriate new pre tag @ 0
	switch info.Kind {
	case ReleaseAlpha:
		newTag.Pre = []semver.PRVersion{{VersionStr: "alpha"}, {VersionNum: 0, IsNum: true}}
	case ReleaseBeta:
		newTag.Pre = []semver.PRVersion{{VersionStr: "beta"}, {VersionNum: 0, IsNum: true}}
	case ReleaseCandidate:
		newTag.Pre = []semver.PRVersion{{VersionStr: "rc"}, {VersionNum: 0, IsNum: true}}
	}

	if semver.Version(newTag).LE(semver.Version(tag)) {
		return newTag, fmt.Errorf("\"new\" version %q actually would be an older version than current %q", newTag.Committish(), tag.Committish())
	}

	return newTag, nil
}

// nextFinalVersion computes the next "final" release given the current one and
// the desired (or lack thereof) to go to v1.0.0.
func (c ChangeLog) nextFinalVersion(current ReleaseTag, pre10 bool) ReleaseTag {
	newTag := semver.Version(current)
	newTag.Pre = nil
	newTag.Build = nil
	switch {
	case len(c.Breaking) > 0:
		if current.Major == 0 && pre10 {
			newTag.IncrementMinor()
		} else {
			newTag.IncrementMajor()
		}
	case len(c.Features) > 0:
		newTag.IncrementMinor()
	// we're doing a new version anyway, so we probably at least need a patch
	default:
		newTag.IncrementPatch()
	}
	return ReleaseTag(newTag)
}

// Changes computes the changelog from last release TO HEAD, returning both the
// changelog and the last release used.
func Changes(gitImpl git.Git, branch *ReleaseBranch) (log ChangeLog, since git.Committish, err error) {
	since, err = CurrentVersion(gitImpl, branch)
	if err != nil {
		return ChangeLog{}, nil, err
	}

	changes, err := ChangesSince(gitImpl, *branch, since)
	return changes, since, err
}

// ChangesSince computes the changelog from the given point to HEAD.
func ChangesSince(gitImpl git.Git, branch ReleaseBranch, since git.Committish) (ChangeLog, error) {
	golog.Printf("finding changes since %q", since.Committish())

	commitsRaw, err := gitImpl.MergeCommitsBetween(since, branch)
	if err != nil {
		return ChangeLog{}, fmt.Errorf("unable to list commits since %s on branch %q: %w", since.Committish(), branch, err)
	}

	log := ChangeLog{}

	// do this parser-style
	commitLines := strings.Split(commitsRaw, "\n")
	lines := &lineReader{lines: commitLines}
	for lines.more() {
		var commit, prNumber, fork string
		if !lines.expectScanf("commit %s", &commit) {
			// skip terminating blank line, and others
			// basically, just get to the next known good state
			if lines.line() != "" {
				golog.Printf("ignoring seemly non-commit line %q", lines.line())
			}
			continue
		}
		if !lines.expectScanf("Merge pull request #%s from %s", &prNumber, &fork) {
			// might be one of the mistakes that got into our history, just
			// bail till the next commit they look like `Merge branch 'BR'`,
			// generally
			golog.Printf("skipping non-official merge commit (%q) with title %q", commit, lines.line())
			continue
		}
		if !lines.expectBlank() {
			golog.Printf("got unexpected non-blank line %q, skipping till next commit", lines.line())
			continue
		}
		if !lines.next() {
			break
		}
		log.entryFromCommit(prNumber, lines.line())
	}

	return log, nil
}

// lineReader helps parsing line-by-line data, like rev-list output.
// start by setting lines.
type lineReader struct {
	lines []string
	cur   string
}

// next loads the next line, returning false if none are available.
func (l *lineReader) next() bool {
	if len(l.lines) == 0 {
		l.cur = ""
		return false
	}
	l.cur = l.lines[0]
	l.lines = l.lines[1:]
	return true
}

// more checks if the next call to next would return true.
func (l *lineReader) more() bool {
	return len(l.lines) > 0
}

// line grabs the current line.
func (l *lineReader) line() string {
	return l.cur
}

// expectScanf loads a new line and scans it according to the supplied args,
// returning false if it didn't scan or no lines were available.
func (l *lineReader) expectScanf(fmtStr string, args ...interface{}) bool {
	if !l.next() {
		return false
	}
	n, err := fmt.Sscanf(l.cur, fmtStr, args...)
	return err == nil && n == len(args)
}

// expectScanf loads a new line and scans it according to the supplied args,
// returning false if it's not blank or no lines were available.
func (l *lineReader) expectBlank() bool {
	if !l.next() {
		return false
	}
	return l.cur == ""
}

// IsPreReleaseToFinal figures out if we're going from a pre-release
// version to a final version.  If true, current is guaranteed to be
// a ReleaseTag.
func IsPreReleaseToFinal(current git.Committish, next ReleaseTag) bool {
	if len(next.Pre) != 0 {
		return false
	}

	tag, isTag := current.(ReleaseTag)
	if !isTag {
		return false
	}

	return len(tag.Pre) != 0
}

// ClosestFinal finds the "closest" previous final release.  For example, given
// `v0.7.0-rc.3`, the closest final release might be `v0.6.3`.
func ClosestFinal(gitImpl git.Git, current ReleaseTag) (*ReleaseTag, error) {
	currentFinal := semver.Version(current)
	currentFinal.Pre = nil

	toCheckTag := semver.Version(current)
	var toCheck git.Committish = current
	for len(toCheckTag.Pre) != 0 || toCheckTag.EQ(currentFinal) {

		tagRaw, err := gitImpl.ClosestTag(git.SomeCommittish(toCheck.Committish() + "~1"))
		if err != nil {
			return nil, err
		}
		latestTag, err := parseReleaseTag(tagRaw)
		if err != nil {
			golog.Printf("skipping non-release tag %q: %v", string(tagRaw), err)
			toCheck = git.SomeCommittish(string(tagRaw))
			continue
		}
		toCheck = latestTag
		toCheckTag = semver.Version(*latestTag)
	}

	if toCheckTag.EQ(semver.Version(current)) || len(toCheckTag.Pre) != 0 {
		return nil, fmt.Errorf("unable to locate previous final release, just found current one")
	}

	tag := ReleaseTag(toCheckTag)
	return &tag, nil
}
