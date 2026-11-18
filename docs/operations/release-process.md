# Release process

## Version rules

Intra uses Semantic Versioning after `v0.1.0`.

- Patch: fixes with no policy, file-format, or CLI contract break.
- Minor: backward-compatible commands, services, or policy fields.
- Major: breaking policy, grant, audit, API, or storage contracts.

Pre-1.0 releases may still break contracts in a minor release when the
changelog names the break and migration path.

## Before a release

1. Confirm [ROADMAP](../../ROADMAP.md) and
   [implementation status](../IMPLEMENTATION-STATUS.md) match reality.
2. Update [VERSION](../../VERSION) and `internal/version/version.go`.
3. Move release notes from `Unreleased` into a dated
   [CHANGELOG](../../CHANGELOG.md) section.
4. Run `./scripts/verify.sh`.
5. Run the named manual release lane:
   - Real Mac Touch ID approval
   - Mac sleep and wake during an active grant
   - Fresh-machine onboarding timing when making an ease claim
6. Scan staged files for secrets.
7. Check `git status --short` and stage files by name.

## Tag and publish

Only a maintainer may create a release tag or publish artifacts.

```sh
git tag -a vX.Y.Z -m "Intra vX.Y.Z"
git push origin main
git push origin vX.Y.Z
```

Publish checksums for every binary or image. State clearly whether each
artifact was built from the tagged commit.

## After release

- Confirm the GitHub release points to the intended tag.
- Confirm `intra version` on a clean download prints the release version.
- Confirm the changelog compare links resolve.
- Start a new `Unreleased` changelog section.
- Record failed or skipped manual checks by name. Never call them passed.

## Update rule

Client updates download and test first. The owner then approves the
apply step. No unattended update may change guard behavior.
