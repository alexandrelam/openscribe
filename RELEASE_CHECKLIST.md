# Release Checklist

Use this checklist when preparing a new release of OpenScribe.

## Pre-Release

- [ ] All features for this release are complete and merged
- [ ] All tests pass locally: `make test`
- [ ] Code is formatted: `make fmt`
- [ ] Linter passes: `make lint` (or warnings are acceptable)
- [ ] Build succeeds for both architectures:
  ```bash
  make build-all
  ```
- [ ] Manual testing completed:
  - [ ] `openscribe setup` works
  - [ ] `openscribe start` works with hotkey detection
  - [ ] Audio recording works
  - [ ] Transcription works
  - [ ] Auto-paste works
  - [ ] Config commands work
  - [ ] Logs commands work
- [ ] Documentation is up to date:
  - [ ] README.md reflects current features
  - [ ] PLAN.md phase checkboxes updated
  - [ ] Any new features documented
- [ ] Git working tree is clean: `git status`

## Creating the Release

- [ ] Decide on version number (semantic versioning):
  - Format: `v<major>.<minor>.<patch>`
  - Examples: `v0.1.0`, `v0.2.0`, `v1.0.0`
  - Patch: Bug fixes only
  - Minor: New features (backward compatible)
  - Major: Breaking changes

- [ ] Create git tag:
  ```bash
  git tag v0.1.0
  ```

- [ ] Push tag to trigger release workflow:
  ```bash
  git push origin v0.1.0
  ```

- [ ] Monitor GitHub Actions:
  - [ ] Go to Actions tab on GitHub
  - [ ] Watch the release workflow
  - [ ] Ensure it completes successfully

- [ ] Verify the GitHub release:
  - [ ] Go to Releases tab
  - [ ] Release should exist for the new version
  - [ ] All 4 artifacts present:
    - `openscribe-darwin-arm64.tar.gz`
    - `openscribe-darwin-amd64.tar.gz`
    - `openscribe-darwin-arm64.tar.gz.sha256`
    - `openscribe-darwin-amd64.tar.gz.sha256`
  - [ ] Release notes look good

## First-Time Homebrew Setup (only for v0.1.0)

- [ ] Create `homebrew-openscribe` repository on GitHub
- [ ] Clone the repository locally:
  ```bash
  git clone git@github.com:alexandrelam/homebrew-openscribe.git
  ```
- [ ] Set up directory structure:
  ```bash
  cd homebrew-openscribe
  mkdir -p Formula
  ```
- [ ] Copy formula template:
  ```bash
  cp /path/to/openscribe/homebrew/openscribe.rb Formula/openscribe.rb
  ```
- [ ] Create README.md in the tap repository
- [ ] Continue to "Updating the Homebrew Formula" section

## Updating the Homebrew Formula

- [ ] Get the new SHA256 checksums:
  ```bash
  # For ARM64
  curl -sL https://github.com/alexandrelam/openscribe/releases/download/v0.1.0/openscribe-darwin-arm64.tar.gz.sha256

  # For AMD64
  curl -sL https://github.com/alexandrelam/openscribe/releases/download/v0.1.0/openscribe-darwin-amd64.tar.gz.sha256
  ```

- [ ] Update the formula using the script:
  ```bash
  ./scripts/update-formula.sh 0.1.0 /path/to/homebrew-openscribe
  ```

  OR manually edit `Formula/openscribe.rb` with:
  - New version number
  - Updated SHA256 checksums
  - Updated URLs

- [ ] Test the formula locally:
  ```bash
  brew audit --strict openscribe
  ```

- [ ] Commit and push to tap repository:
  ```bash
  cd /path/to/homebrew-openscribe
  git add Formula/openscribe.rb
  git commit -m "Update openscribe to v0.1.0"
  git push origin main
  ```

## Testing the Release

- [ ] Test Homebrew installation:
  ```bash
  # Tap the repository (if not already tapped)
  brew tap alexandrelam/openscribe

  # Install
  brew install openscribe

  # Verify version
  openscribe version
  ```

- [ ] Test full workflow:
  ```bash
  openscribe setup
  openscribe config --show
  openscribe models list
  # ... test other commands
  ```

- [ ] Uninstall and reinstall to test clean installation:
  ```bash
  brew uninstall openscribe
  brew untap alexandrelam/openscribe
  brew tap alexandrelam/openscribe
  brew install openscribe
  ```

- [ ] Test on both architectures (if possible):
  - [ ] Apple Silicon (ARM64)
  - [ ] Intel (x86_64)

## Post-Release

- [ ] Update main repository README (if needed)
- [ ] Announce the release:
  - [ ] GitHub Discussions
  - [ ] Social media (if applicable)
  - [ ] Email list (if applicable)
- [ ] Monitor for issues:
  - [ ] Check GitHub Issues
  - [ ] Respond to installation problems
  - [ ] Note any bugs for next release

## For Emergency Hotfixes

If you need to quickly fix a critical bug:

1. Create a branch from the release tag:
   ```bash
   git checkout -b hotfix-v0.1.1 v0.1.0
   ```

2. Make the fix and commit:
   ```bash
   git commit -am "Fix critical bug XYZ"
   ```

3. Create new tag and release:
   ```bash
   git tag v0.1.1
   git push origin v0.1.1
   git push origin hotfix-v0.1.1
   ```

4. Merge hotfix back to main:
   ```bash
   git checkout main
   git merge hotfix-v0.1.1
   git push origin main
   ```

5. Update Homebrew formula with new version

## Version History

Track your releases here:

- [ ] v0.1.0 - Initial release (YYYY-MM-DD)
- [ ] v0.1.1 - Bug fixes (YYYY-MM-DD)
- [ ] v0.2.0 - New features (YYYY-MM-DD)

---

## Quick Commands Reference

```bash
# Build and test locally
make clean
make build-all
./bin/openscribe version

# Create and push release tag
git tag v0.1.0
git push origin v0.1.0

# Update formula (after release is created)
./scripts/update-formula.sh 0.1.0 /path/to/homebrew-openscribe

# Test installation
brew tap alexandrelam/openscribe
brew install openscribe
openscribe version
```

---

## Troubleshooting

### GitHub Actions fails
- Check the workflow logs on GitHub
- Verify Go version in workflow matches go.mod
- Test local build with `make build-all`

### SHA256 mismatch
- Re-download the .sha256 files
- Make sure you're using the correct version number
- Verify the release assets exist on GitHub

### Formula won't install
- Run `brew audit --strict openscribe` to check for errors
- Verify Ruby syntax in openscribe.rb
- Check that URLs point to correct release

---

## Notes

- Always test locally before releasing
- Keep semantic versioning consistent
- Update documentation before tagging
- Communicate breaking changes clearly
- Monitor issues after release
