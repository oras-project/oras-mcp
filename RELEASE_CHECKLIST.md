# Release Checklist

## Overview

This document is for maintainers and describes the checklist to publish an `oras-mcp` release through the GitHub release page. After releasing, an updated multi-architecture container image will be pushed to `ghcr.io/oras-project/oras-mcp`.

## Release Process

1. Determine a [SemVer2](https://semver.org/)-valid version prefixed with `v` (for example, `v1.0.0` or `v1.1.0-rc.1`).
2. Update `internal/version/version.go` with the new version and open a pull request that serves as the release vote. The commit in that PR is the one that will be tagged.
3. Collect approvals and make sure CI for the PR is green. Once the vote passes, keep a note of the exact commit SHA that was reviewed.
4. Merge the pull request using **Create a merge commit** so the voted commit stays intact in the target branch.
5. In a clean working tree, check out the voted commit (not the merge commit) and create the release tag, then push it:
   - `git tag vX.Y.Z <commit-sha>`
   - `git push origin vX.Y.Z`
6. (Optional) Cut off a `release-<major>.<minor>` branch on the tagged commit, but only when you are shipping a new minor version.
7. Watch the `release-github.yml` and `release-ghcr.yml` workflows triggered by the new tag and wait for them to succeed.
8. Validate the published image on GHCR by pulling and running a quick smoke test such as `docker run --rm ghcr.io/oras-project/oras-mcp:vX.Y.Z serve --help`.
9. Polish the release notes (either in the PR description or a shared doc) and copy them into the GitHub release draft for `vX.Y.Z`.
10. Publish the GitHub release once the notes are ready and validation is complete.
11. Announce the release in the [#oras](https://cloud-native.slack.com/archives/CJ1KHJM5Z) Slack channel and close out any tracking items.
