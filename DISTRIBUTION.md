# OpenScribe Distribution Guide

This guide covers the complete process for distributing OpenScribe via Homebrew, from creating releases to updating the formula.

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Initial Setup](#initial-setup)
4. [Creating a Release](#creating-a-release)
5. [Setting Up the Homebrew Tap](#setting-up-the-homebrew-tap)
6. [Updating for New Releases](#updating-for-new-releases)
7. [Testing](#testing)
8. [Troubleshooting](#troubleshooting)

---

## Overview

OpenScribe uses a **Homebrew tap** for distribution. This approach provides:

- **Easy installation**: Users can install with `brew tap` and `brew install`
- **Automatic updates**: Users get updates via `brew update` and `brew upgrade`
- **Dependency management**: Homebrew handles whisper-cpp dependency
- **Multi-architecture support**: Separate binaries for ARM64 and x86_64

### Distribution Flow

```
1. Tag a new version (e.g., v0.1.0)
   ↓
2. GitHub Actions builds binaries and creates release
   ↓
3. Update Homebrew formula with new checksums
   ↓
4. Users install/update via Homebrew
```

---

## Prerequisites

Before setting up distribution, ensure you have:

1. **GitHub repository** with proper permissions
2. **GitHub Actions** enabled on the repository
3. **Git tags** for versioning (semantic versioning recommended)
4. **Homebrew** installed locally for testing

---

## Initial Setup

### Step 1: Create the Homebrew Tap Repository

A Homebrew tap is a separate GitHub repository that contains your formula.

1. **Create a new repository** on GitHub named `homebrew-openscribe`
   - The `homebrew-` prefix is **required** by Homebrew convention
   - Repository should be public
   - Initialize with a README

2. **Clone the repository**:
   ```bash
   git clone git@github.com:alexandrelam/homebrew-openscribe.git
   cd homebrew-openscribe
   ```

3. **Create the Formula directory**:
   ```bash
   mkdir -p Formula
   ```

4. **Copy the formula template**:
   ```bash
   # From the openscribe repository
   cp homebrew/openscribe.rb /path/to/homebrew-openscribe/Formula/openscribe.rb
   ```

5. **Create a README** for the tap:
   ```bash
   cat > README.md << 'EOF'
   # OpenScribe Homebrew Tap

   Official Homebrew tap for OpenScribe - Real-time speech transcription CLI for macOS.

   ## Installation

   ```bash
   brew tap alexandrelam/openscribe
   brew install openscribe
   openscribe setup
   ```

   ## About

   OpenScribe is a free, open-source speech-to-text application for macOS that provides:
   - 100% offline processing
   - Universal compatibility across all macOS apps
   - Hotkey-activated transcription
   - Multiple Whisper model sizes

   For more information, visit: https://github.com/alexandrelam/openscribe

   ## Support

   - Issues: https://github.com/alexandrelam/openscribe/issues
   - Documentation: https://github.com/alexandrelam/openscribe
   EOF
   ```

6. **Commit and push**:
   ```bash
   git add Formula/openscribe.rb README.md
   git commit -m "Initial formula for openscribe"
   git push origin main
   ```

### Step 2: Configure GitHub Actions

The GitHub Actions workflow (`.github/workflows/release.yml`) is already set up in the main repository. It will:

- Trigger on git tags matching `v*` (e.g., `v0.1.0`)
- Build binaries for ARM64 and x86_64
- Create tar.gz archives
- Generate SHA256 checksums
- Create a GitHub release with all artifacts

No additional configuration is needed!

---

## Creating a Release

### Step 1: Prepare the Release

1. **Ensure all changes are committed**:
   ```bash
   git status  # Should show clean working tree
   ```

2. **Update version in code** (if needed):
   - The Makefile automatically uses git tags for versioning
   - No manual version updates needed

3. **Test locally**:
   ```bash
   make build
   ./bin/openscribe version
   ./bin/openscribe --help
   ```

### Step 2: Create and Push a Tag

1. **Create a git tag** with semantic versioning:
   ```bash
   git tag v0.1.0
   ```

   Versioning guidelines:
   - `v0.1.0` - Initial release
   - `v0.1.1` - Bug fixes
   - `v0.2.0` - New features (backward compatible)
   - `v1.0.0` - Stable release

2. **Push the tag** to GitHub:
   ```bash
   git push origin v0.1.0
   ```

### Step 3: Monitor GitHub Actions

1. Go to your repository on GitHub
2. Click the **Actions** tab
3. You should see a workflow running for the new tag
4. Wait for the workflow to complete (~5-10 minutes)

### Step 4: Verify the Release

1. Go to **Releases** tab on GitHub
2. You should see a new release for `v0.1.0` with:
   - Release notes (auto-generated)
   - `openscribe-darwin-arm64.tar.gz`
   - `openscribe-darwin-amd64.tar.gz`
   - `openscribe-darwin-arm64.tar.gz.sha256`
   - `openscribe-darwin-amd64.tar.gz.sha256`

---

## Setting Up the Homebrew Tap

Now that the release is created, update the Homebrew formula with the correct checksums.

### Option A: Using the Update Script (Recommended)

1. **Run the update script**:
   ```bash
   ./scripts/update-formula.sh 0.1.0 /path/to/homebrew-openscribe
   ```

   This script will:
   - Fetch SHA256 checksums from the release
   - Update the formula with correct version and checksums
   - Show next steps for testing and publishing

2. **Follow the output instructions** to commit and push

### Option B: Manual Update

1. **Download the checksums**:
   ```bash
   curl -sL https://github.com/alexandrelam/openscribe/releases/download/v0.1.0/openscribe-darwin-arm64.tar.gz.sha256
   curl -sL https://github.com/alexandrelam/openscribe/releases/download/v0.1.0/openscribe-darwin-amd64.tar.gz.sha256
   ```

2. **Edit the formula** at `Formula/openscribe.rb`:
   ```ruby
   version "0.1.0"

   if Hardware::CPU.arm?
     url "https://github.com/alexandrelam/openscribe/releases/download/v0.1.0/openscribe-darwin-arm64.tar.gz"
     sha256 "actual_arm64_sha256_here"
   else
     url "https://github.com/alexandrelam/openscribe/releases/download/v0.1.0/openscribe-darwin-amd64.tar.gz"
     sha256 "actual_amd64_sha256_here"
   end
   ```

3. **Commit and push**:
   ```bash
   cd /path/to/homebrew-openscribe
   git add Formula/openscribe.rb
   git commit -m "Update openscribe to v0.1.0"
   git push origin main
   ```

### Test the Installation

1. **Tap the repository** (if not already tapped):
   ```bash
   brew tap alexandrelam/openscribe
   ```

2. **Install OpenScribe**:
   ```bash
   brew install openscribe
   ```

3. **Verify installation**:
   ```bash
   openscribe version
   openscribe --help
   ```

4. **Test the full workflow**:
   ```bash
   openscribe setup
   openscribe config --show
   ```

---

## Updating for New Releases

When you release a new version, follow these steps:

### 1. Create the New Release

```bash
# In the main openscribe repository
git tag v0.2.0
git push origin v0.2.0
```

Wait for GitHub Actions to complete and create the release.

### 2. Update the Formula

```bash
# Use the update script
./scripts/update-formula.sh 0.2.0 /path/to/homebrew-openscribe
```

Or manually update as described in [Setting Up the Homebrew Tap](#option-b-manual-update).

### 3. Commit and Push

```bash
cd /path/to/homebrew-openscribe
git add Formula/openscribe.rb
git commit -m "Update openscribe to v0.2.0"
git push origin main
```

### 4. Users Update

Users can now update to the new version:

```bash
brew update
brew upgrade openscribe
```

---

## Testing

### Local Formula Testing

Before publishing, test the formula locally:

```bash
# Audit the formula for issues
brew audit --strict --online openscribe

# Install from source
brew install --build-from-source openscribe

# Test the binary
openscribe version

# Uninstall
brew uninstall openscribe
```

### Testing on a Clean System

For thorough testing:

1. **Use a virtual machine or fresh macOS install**
2. **Install Homebrew** (if not present)
3. **Tap and install**:
   ```bash
   brew tap alexandrelam/openscribe
   brew install openscribe
   ```
4. **Run through the full workflow**:
   ```bash
   openscribe setup
   openscribe start
   ```

### Automated Testing

The formula includes a test block:

```ruby
test do
  assert_match version.to_s, shell_output("#{bin}/openscribe version")
end
```

This runs when users install with `--build-from-source` and during `brew audit`.

---

## Troubleshooting

### Release Assets Not Found

**Problem**: Formula can't download release assets.

**Solutions**:
1. Verify the release exists on GitHub
2. Check that asset names match the formula URLs
3. Ensure the release is published (not draft)
4. Wait a few minutes for CDN propagation

### SHA256 Mismatch

**Problem**: `Error: SHA256 mismatch` during installation.

**Solutions**:
1. Re-download the SHA256 files from the release
2. Verify you're using the correct version number
3. Clear Homebrew cache: `brew cleanup`
4. Force reinstall: `brew reinstall openscribe`

### Formula Syntax Errors

**Problem**: `Error: Invalid formula` during installation.

**Solutions**:
1. Run `brew audit openscribe` to check for errors
2. Verify Ruby syntax in the formula file
3. Check indentation (use spaces, not tabs)
4. Compare with the template in `homebrew/openscribe.rb`

### Build Failures

**Problem**: GitHub Actions workflow fails.

**Solutions**:
1. Check the Actions logs on GitHub
2. Verify Go version in workflow matches `go.mod`
3. Ensure all dependencies are available
4. Test local build with `make build-all`

### whisper-cpp Dependency Not Found

**Problem**: Users can't install due to missing whisper-cpp.

**Solutions**:
1. Ensure whisper-cpp is available in Homebrew
2. Users should run: `brew install whisper-cpp`
3. The formula already lists this as a dependency

---

## Best Practices

### Versioning

- **Use semantic versioning**: `v1.0.0`, `v1.1.0`, `v1.0.1`
- **Tag format**: Always use `v` prefix (e.g., `v0.1.0`)
- **Breaking changes**: Bump major version
- **New features**: Bump minor version
- **Bug fixes**: Bump patch version

### Release Notes

- Let GitHub auto-generate release notes
- Add manual notes for breaking changes
- Include upgrade instructions if needed
- Link to documentation for new features

### Formula Maintenance

- **Test before publishing**: Always test locally
- **Keep caveats updated**: Update post-install messages as needed
- **Document dependencies**: List all required dependencies
- **Version constraints**: Specify minimum macOS version if needed

### Communication

- **Announce releases**: Post on GitHub Discussions
- **Update documentation**: Keep README in sync
- **Notify users**: For breaking changes, create an issue/discussion
- **Changelog**: Maintain a CHANGELOG.md in the main repo

---

## Resources

- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Homebrew Tap Documentation](https://docs.brew.sh/Taps)
- [Creating and Maintaining a Tap](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
- [GitHub Actions for Go](https://docs.github.com/en/actions/guides/building-and-testing-go)
- [Semantic Versioning](https://semver.org/)

---

## Quick Reference

### Common Commands

```bash
# Build locally
make build

# Build all architectures
make build-all

# Create and push a tag
git tag v0.1.0 && git push origin v0.1.0

# Update formula
./scripts/update-formula.sh 0.1.0 /path/to/homebrew-openscribe

# Test formula locally
brew audit --strict openscribe
brew install --build-from-source openscribe

# Users install
brew tap alexandrelam/openscribe
brew install openscribe

# Users update
brew update
brew upgrade openscribe
```

### File Locations

```
Main Repository:
├── .github/workflows/release.yml    # GitHub Actions for releases
├── homebrew/openscribe.rb           # Formula template
├── homebrew/README.md               # Homebrew setup guide
├── scripts/update-formula.sh        # Formula update automation
└── Makefile                         # Build targets

Tap Repository:
└── Formula/openscribe.rb            # Published formula
```

---

## Next Steps

1. ✅ Create the `homebrew-openscribe` repository
2. ✅ Set up the initial formula
3. ✅ Create your first release (v0.1.0)
4. ✅ Update the formula with checksums
5. ✅ Test the installation
6. ✅ Announce to users!

Happy distributing! 🎉
