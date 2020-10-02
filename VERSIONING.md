# Versioning and Branching in KubeBuilder projects

*NB*: this applies to [KubeBuilder](https://sigs.k8s.io/kubebuilder),
[controller-runtime](https://sigs.k8s.io/controller-runtime),
[controller-tools](https://sigs.k8s.io/controller-tools), this repo, and
any other projects under the purview of KubeBuilder.  These projects will
be known in the doc below as "KubeBuilder projects".

Individual projects may have more specific rules or guidelines.

## TL;DR:

### Users

- We follow [Semantic Versioning (semver)](https://semver.org)
- Use releases with your dependency management to ensure that you get
  compatible code

- The master branch contains all the latest code, some of which may break
  compatibility.

- We support go modules as our official versioning mechanism.

### Contributors

- All code PR must be labeled with :bug: (patch fixes), :sparkles:
  (backwards-compatible features), or :warning: (breaking changes)

- Breaking changes will find their way into the next major release, other
  changes will go into an semi-immediate patch or minor release

- Please *try* to avoid breaking changes when you can.  They make users
  face difficult decisions ("when do I go through the pain of
  upgrading?"), and make life hard for maintainers and contributors
  (dealing with differences on stable branches).

### Mantainers

Don't be lazy, read the rest of this doc :-)

## Overview

KubeBuilder projects follow [Semantic Versioning](https://semver.org). I'd
recommend reading the aforementioned link if you're not familiar, but
essentially, for any given release X.Y.Z:

- an X (*major*) release indicates a set of backwards-compatible code.
  Changing X means there's a breaking change.

- a Y (*minor*) release indicates a minimum feature set.  Changing Y means
  the addition of a backwards-compatible feature.

- a Z (*patch*) release indicates minimum set of bugfixes.  Changing
  Z means a backwards-compatible change that doesn't add functionality.

*NB*: If the major release is `0`, any minor release may contain breaking
changes.

For library projects (**controller-runtime**), these guarantees extend to
all code exposed in public APIs of controller-runtime. This includes code
both in controller-runtime itself, *plus types from dependencies in public
APIs*. Types and functions not in public APIs are not considered part of
the guarantee.

For CLI projects (**controller-tools**, **KubeBuilder**) these guarantees
cover the interface presented to the user by the CLI instead of the code
itself.  Projects that are both follow both requirements.

In order to easily maintain the guarantees, we have a couple of processes
that we follow.

## Go Module Paths

All projects (regardless of whether they're CLI or library projects)
should follow the [Go versioning path rules][go-path-rules].

[go-path-rules]: https://golang.org/ref/mod#major-version-suffixes

## Branches

controller-runtime contains two types of branches: the *master* branch and
*release-X* branches.

The *master* branch is where development happens.  All the latest and
greatest code, including breaking changes, happens on master.

The *release-X* branches contain stable, backwards compatible code.  Every
major (X) release, a new such branch is created.  It is from these
branches that minor and patch releases are tagged.  If some cases, it may
be necessary open PRs for bugfixes directly against stable branches, but
this should generally not be the case.

The maintainers are responsible for updating the contents of this branch;
generally, this is done just before a release using release tooling that
filters and checks for changes tagged as breaking (see below).

Individual projects have different policies on how many release branches
may be actively supported at any particular time.

## Kubernetes Version Compatibility

This may vary slightly project to project, but we tend to guarantee
**Kubernetes REST API support** for N-3 or N-4 versions of Kubernetes,
where `N` is the current latest release (not including pre-releases).

Note that this does not guarantee **dependency support**.  For instance,
while a given version of `controller-runtime` may work with Kubernetes
1.16 through 1.19, it may only compile with `client-go@v0.19.0`.

### Tooling

Tooling for releases lives in the [kubebuilder-release-tools
repository][rel-tools]:

* `go run sigs.k8s.io/kubebuilder-release-tools/notes` from a `release-X`
  (or `release-0.Y`) branch will generate release notes based on the emoji
  described below.
  
  You can generally paste the stdout output directly into the release,
  unless the tool mentions needing manual edits.

* the [GitHub actions][action] in this repo will verify PRs using the
  verifier code from [verify](/verify).  If you want to add new checks,
  you can do it there.

[rel-tools]: https://sigs.k8s.io/kubebuilder-release-tools

[action]: /action.yml

## PR Process

Every PR should be annotated with an icon indicating whether it's
a:

- Breaking change: :warning: (`:warning:`)
- Non-breaking feature: :sparkles: (`:sparkles:`)
- Patch fix: :bug: (`:bug:`)
- Docs: :book: (`:book:`)
- Infra/Tests/Other: :seedling: (`:seedling:`)
- No release note: :ghost: (`:ghost:`)

Use :ghost: (no release note) only for the PRs that change or revert unreleased
changes, which don't deserve a release note. Please don't abuse it.

You can also use the equivalent emoji directly, since GitHub doesn't
render the `:xyz:` aliases in PR titles.

Individual commits should not be tagged separately, but will generally be
assumed to match the PR. For instance, if you have a bugfix in with
a breaking change, it's generally encouraged to submit the bugfix
separately, but if you must put them in one PR, you should mark the whole
PR as breaking.

### Commands and Workflow

KubeBuilder projects follow the standard Kubernetes workflow: any PR
needs `lgtm` and `approved` labels, PRs authors must have signed the CNCF
CLA, and PRs must pass the tests before being merged.  See [the
contributor
docs](https://github.com/kubernetes/community/blob/master/contributors/guide/pull-requests.md#the-testing-and-merge-workflow)
for more info.

We use the same priority and kind labels as Kubernetes.  See the labels
tab in GitHub for the full list.

The standard Kubernetes comment commands should work in KubeBuilder
projects. See [Prow](https://prow.k8s.io/command-help) for a command
reference.

## Release Process

Minor and patch releases are generally done immediately after a feature or
bugfix is landed, or sometimes a series of features tied together.

Minor releases will only be tagged on the *most recent* major release
branch, except in exceptional circumstances.  Patches will be backported
to maintained stable versions, as needed.

Major releases are generally done shortly after a breaking change is
merged -- once a breaking change is merged, the next release *must* be
a major revision. We don't intend to have a lot of these, so we may put
off merging breaking PRs until a later date.

### Exact Steps

Follow the release-specific steps below, then follow the general steps
after that.

#### Minor and patch releases

1. Update the release-X branch with the latest set of changes by calling
   `git rebase master` from the release branch.

#### Major releases

1. Create a new release branch named `release-X` (where `X` is the new
   version) off of master.

#### General

2. Generate release notes using the release note tooling.

3. Add a release for the project on GitHub, using those release notes,
   with a title of `vX.Y.Z`.

4. Announce the release in `#kubebuilder` on Slack with a pinned message.

You may also need to update dependent projects (e.g. updates to
controller-tools generally mean an update to the KubeBuilder book
tooling).

### Breaking Changes

Try to avoid breaking changes.  They make life difficult for users, who
have to rewrite their code when they eventually upgrade, and for
maintainers/contributors, who have to deal with differences between master
and stable branches.

That being said, we'll occasionally want to make breaking changes. They'll
be merged onto master, and will then trigger a major release (see [Release
Process](#release-process)).  Because breaking changes induce a major
revision, the maintainers may delay a particular breaking change until
a later date when they are ready to make a major revision with a few
breaking changes.

If you're going to make a breaking change, please make sure to explain in
detail why it's helpful.  Is it necessary to cleanly resolve an issue?
Does it improve API ergonomics?

Maintainers should treat breaking changes with caution, and evaluate
potential non-breaking solutions (see below).

Note that API breakage in public APIs due to dependencies will trigger
a major revision in library projects, so you may occasionally need to have
a major release anyway, due to changes in libraries like
`k8s.io/client-go` or `k8s.io/apimachinery`.

*NB*: Pre-1.0 releases treat breaking changes a bit more lightly.  We'll
still consider carefully, but the pre-1.0 timeframe is useful for
converging on a ergonomic API.

#### Avoiding breaking changes

##### Solutions to avoid

- **Confusingly duplicate methods, functions, or variables.**

  For instance, suppose we have an interface method `List(ctx
  context.Context, options *ListOptions, obj runtime.Object) error`, and
  we decide to switch it so that options come at the end, parametrically.
  Adding a new interface method `ListParametric(ctx context.Context, obj
  runtime.Object, options... ListOption)` is probably not the right
  solution:

   - Users will intuitively see `List`, and use that in new projects, even
     if it's marked as deprecated.

   - Users who don't notice the deprecation may be confused as to the
     difference between `List` and `ListParametric`.

   - It's not immediately obvious in isolation (e.g. in surrounding code)
     why the method is called `ListParametric`, and may cause confusion
     when reading code that makes use of that method.

  In this case, it may be better to make the breaking change, and then
  eventually do a major release.

## Why don't we...

### Use "next"-style branches

Development branches:

- don't win us much in terms of maintenance in the case of breaking
  changes (we still have to merge/manage multiple branches for development
  and stable)

- can be confusing to contributors, who often expect master to have the
  latest changes.

### Never break compatibility

Never doing a new major release could be an admirable goal, but gradually
leads to API cruft.

Since one of the goals of KubeBuilder projects is to be friendly and
intuitive, we want to avoid too much API cruft over time, and occasional
breaking changes in major releases help accomplish that goal.

Furthermore, our dependency on Kubernetes libraries makes this difficult
(see below)

### Always assume we've broken compatibility

*a.k.a. k8s.io/client-go style*

While this makes life easier (a bit) for maintainers, it's problematic for
users.  While breaking changes arrive sooner, upgrading becomes very
painful.

Furthermore, we still have to maintain stable branches for bugfixes, so
the maintenance burden isn't lessened by a ton.

### Extend compatibility guarantees to all dependencies

This is very difficult with the number of Kubernetes dependencies we have.
Kubernetes dependencies tend to either break compatibility every major
release (e.g. k8s.io/client-go, which loosely follows semver), or at
a whim (many other Kubernetes libraries).

If we limit to the few objects we expose, we can better inform users about
how *KubeBuilder projects themselves* have changed in a given release.
Then, users can make informed decisions about how to proceed with any
direct uses of Kubernetes dependencies their controller-runtime-based
application may have.

