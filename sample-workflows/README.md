# Sample GitHub Actions Workflows

These include GitHub Actions Worflows for use in KubeBuilder projects.

## PR Verifier

**File(s)**:

- [verifier.yml](verifier.yml) (`/.github/workflows/verifier.yml`)

This uses the [PR Verifier Action](/action.yml) to verify PR title and
contents according to [the PR guidelines](/VERSIONING.md).

[verifier-action]: /action.yml

## Lint

**File(s)**:

- [lint.yml](lint.yml) (`/.github/workflows/lint.yml`)
- [.golangci.yml](.golangci.yml) (`/.golangci.yml`)

This uses [golangci-lint](https://github.com/golangci/golangci-lint) to
lint our code.

Use the included config file at the root of your repo to configure what
linters run by default (golangci-lint has some strange defaults, and this
config should be used more-or-less for all KubeBuilder projects).
