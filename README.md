# Kubebuilder Release Tools

Release tooling for KubeBuilder projects.

## Release Notes Generation

The [notes](/notes) module contains a framework for generating release
notes from git history using emoji, and the "root" of the module is
a program that makes use of this:

```shell
# generate a final release
$ go run sigs.k8s.io/kubebuilder-release-tools/notes

  generate a beta release
$ go run sigs.k8s.io/kubebuilder-release-tools/notes -r beta
```

## PR Verification GitHub Action

This repository acts as a GitHub action for verifying PR titles match the
[release notes generation requirements](/VERSIONING.md), as well as some
basic descriptiveness checks.  You can use it in your repository by adding
a workflow (e.g. `.github/workflows/verifier.yml`) as such:

```yaml
name: PR Verifier

on:
  # NB: using `pull_request_target` runs this in the context of
  # the base repository, so it has permission to upload to the checks API.
  # This means changes won't kick in to this file until merged onto the
  # main branch.
  pull_request_target:
    types: [opened, edited, reopened, synchronize]

jobs:
  verify:
    runs-on: ubuntu-latest
    name: verify PR contents
    steps:
    - name: Verifier action
      id: verifier
      uses: kubernetes-sigs/kubebuilder-release-tools@v0.2
      with:
        github_token: ${{ secrets.GITHUB_TOKEN }}
```

The code that actually runs lives in [verify](/verify), while
[/verify/pkg/action](/verify/pkg/action) contains a framework for running PR
description checks from GitHub actions & uploading the result via the GitHub
checks API.

This repo itself uses a "live" version of the action that always rebuilds
from the local code (master branch), which lives in 
[.github/actions/verifier](/.github/actions/verifier).

### Updating the action

If you release updates to the action, make sure to tag a new release,
which triggers a build & tag of the docker container referenced by this
action (using Google Cloud Build, pushed as
[gcr.io/kubebuilder/pr-verifier](https://gcr.io/kubebuilder/pr-verifier)).
and then update the corresponding major version tag (either `vX` or
`v0.Y`) by running:

```shell
# where vX is the major version, vX.Y.Z is the release you just tagged,
# and upstream is the remote for this repo itself, NOT your fork.
$ git pull --tags upstream
$ git tag -f vX vX.Y.Z
$ git push upstream refs/tags/vX
```

## KubeBuilder Project Versioning

[VERSIONING.md](/VERSIONING.md) contains the general guidelines for
versioning, releasing, etc for the KubeBuilder projects.

The TL;DR on PR titles is that you must have a *prefix* on your PR title
specifying the kind of change:

- Breaking change: :warning: (`:warning:`)
- Non-breaking feature: :sparkles: (`:sparkles:`)
- Patch fix: :bug: (`:bug:`)
- Docs: :book: (`:book:`)
- Infra/Tests/Other: :seedling: (`:seedling:`)

See [the document](/VERSIONING.md) for more details.

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack](https://kubernetes.slack.com/messages/kubebuilder)
- [Mailing List](https://groups.google.com/forum/#!forum/kubebuilder)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
