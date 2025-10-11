---
name: Release Tracking
about: Track the release process for a new version
title: 'Release v0.X.X'
labels: release
assignees: ''
---

## Release Version

**Version:** v0.X.X

**Release Type:** (Major / Minor / Patch)

**Target Date:** YYYY-MM-DD

---

## Pre-Release Checklist

- [ ] All features complete and merged
- [ ] Tests pass: `make test`
- [ ] Code formatted: `make fmt`
- [ ] Linter passes: `make lint`
- [ ] Build succeeds: `make build-all`
- [ ] Manual testing complete
- [ ] Documentation updated
- [ ] Git working tree clean

---

## Release Creation

- [ ] Version number decided: `v0.X.X`
- [ ] Git tag created: `git tag v0.X.X`
- [ ] Tag pushed: `git push origin v0.X.X`
- [ ] GitHub Actions workflow completed successfully
- [ ] GitHub release created with all artifacts

---

## Homebrew Formula Update

- [ ] SHA256 checksums retrieved
- [ ] Formula updated with `./scripts/update-formula.sh`
- [ ] Formula committed to tap repository
- [ ] Formula tested with `brew audit`

---

## Testing

- [ ] Homebrew installation tested
- [ ] Full workflow tested
- [ ] Tested on Apple Silicon (ARM64)
- [ ] Tested on Intel (x86_64)

---

## Post-Release

- [ ] README updated (if needed)
- [ ] Release announced
- [ ] Issues monitored
- [ ] Version history updated in RELEASE_CHECKLIST.md

---

## Release Notes

**New Features:**
- Feature 1
- Feature 2

**Bug Fixes:**
- Fix 1
- Fix 2

**Breaking Changes:**
- Breaking change 1 (if any)

**Known Issues:**
- Issue 1 (if any)

---

## Links

- GitHub Release: https://github.com/alexandrelam/openscribe/releases/tag/v0.X.X
- Homebrew Formula: https://github.com/alexandrelam/homebrew-openscribe/blob/main/Formula/openscribe.rb

---

## Notes

(Add any additional notes about this release here)
