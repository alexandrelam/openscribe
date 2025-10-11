# Homebrew Distribution Setup - Summary

This document summarizes everything that has been set up for distributing OpenScribe via Homebrew.

## ✅ What's Been Created

### 1. GitHub Actions Workflow
**File:** `.github/workflows/release.yml`

Automatically builds and releases OpenScribe when you push a git tag:
- Builds for macOS ARM64 and x86_64
- Creates tar.gz archives
- Generates SHA256 checksums
- Creates GitHub release with all artifacts

**Triggers on:** Tags matching `v*` (e.g., `v0.1.0`, `v1.0.0`)

### 2. Makefile Build Targets
**File:** `Makefile`

New targets added:
- `make build-darwin-arm64` - Build for Apple Silicon
- `make build-darwin-amd64` - Build for Intel Macs
- `make build-all` - Build both architectures

### 3. Homebrew Formula Template
**File:** `homebrew/openscribe.rb`

Ready-to-use Homebrew formula with:
- Multi-architecture support (ARM64 + x86_64)
- whisper-cpp dependency
- Post-install instructions (caveats)
- Accessibility and microphone permissions guidance
- Version test

### 4. Formula Update Script
**File:** `scripts/update-formula.sh`

Automation script that:
- Fetches SHA256 checksums from GitHub releases
- Updates the formula with new version and checksums
- Provides next-step instructions

**Usage:** `./scripts/update-formula.sh 0.1.0 /path/to/homebrew-openscribe`

### 5. Documentation

#### `homebrew/README.md`
- Detailed Homebrew tap setup guide
- Formula maintenance instructions
- Testing procedures
- Troubleshooting tips
- Resources and best practices

#### `DISTRIBUTION.md`
- Complete distribution guide
- Step-by-step setup instructions
- Release process documentation
- Testing strategies
- Quick reference commands

#### `RELEASE_CHECKLIST.md`
- Pre-release checklist
- Release creation steps
- Homebrew formula update process
- Testing procedures
- Post-release tasks

#### `README.md` (updated)
- Homebrew installation instructions
- Alternative installation methods
- Updated user-facing documentation

### 6. GitHub Issue Template
**File:** `.github/ISSUE_TEMPLATE/release.md`

Template for tracking releases with:
- Pre-release checklist
- Release creation steps
- Testing checklist
- Post-release tasks
- Release notes section

---

## 📋 Next Steps

To actually publish OpenScribe to Homebrew, follow these steps:

### Step 1: Create the Homebrew Tap Repository

1. **Go to GitHub** and create a new repository:
   - Name: `homebrew-openscribe` (the `homebrew-` prefix is required!)
   - Visibility: Public
   - Initialize with README: Yes

2. **Clone it locally:**
   ```bash
   git clone git@github.com:alexandrelam/homebrew-openscribe.git
   cd homebrew-openscribe
   ```

3. **Create Formula directory:**
   ```bash
   mkdir -p Formula
   ```

### Step 2: Create Your First Release

1. **Test everything locally:**
   ```bash
   make clean
   make build-all
   ./bin/openscribe version
   ./bin/openscribe --help
   ```

2. **Create and push a git tag:**
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

3. **Wait for GitHub Actions** to complete (check Actions tab on GitHub)

4. **Verify the release** on GitHub:
   - Go to Releases tab
   - Should see v0.1.0 with 4 artifacts:
     - `openscribe-darwin-arm64.tar.gz`
     - `openscribe-darwin-amd64.tar.gz`
     - `openscribe-darwin-arm64.tar.gz.sha256`
     - `openscribe-darwin-amd64.tar.gz.sha256`

### Step 3: Update the Homebrew Formula

1. **Run the update script:**
   ```bash
   ./scripts/update-formula.sh 0.1.0 /path/to/homebrew-openscribe
   ```

2. **Commit and push** to the tap repository:
   ```bash
   cd /path/to/homebrew-openscribe
   git add Formula/openscribe.rb
   git commit -m "Add openscribe formula v0.1.0"
   git push origin main
   ```

### Step 4: Test the Installation

1. **Tap your repository:**
   ```bash
   brew tap alexandrelam/openscribe
   ```

2. **Install OpenScribe:**
   ```bash
   brew install openscribe
   ```

3. **Verify it works:**
   ```bash
   openscribe version
   openscribe --help
   openscribe setup
   ```

### Step 5: Announce!

Update your main README (already done!) and let users know they can install via Homebrew:

```bash
brew tap alexandrelam/openscribe
brew install openscribe
```

---

## 🔄 For Future Releases

When you want to release a new version:

1. **Create a new tag:**
   ```bash
   git tag v0.2.0
   git push origin v0.2.0
   ```

2. **Wait for GitHub Actions** to build and create the release

3. **Update the formula:**
   ```bash
   ./scripts/update-formula.sh 0.2.0 /path/to/homebrew-openscribe
   ```

4. **Commit and push** to the tap repository

5. **Users update with:**
   ```bash
   brew update
   brew upgrade openscribe
   ```

---

## 📁 File Structure

Here's what was added to your repository:

```
openscribe/
├── .github/
│   ├── workflows/
│   │   └── release.yml              # GitHub Actions for releases
│   └── ISSUE_TEMPLATE/
│       └── release.md               # Release tracking template
├── homebrew/
│   ├── openscribe.rb                # Homebrew formula template
│   ├── README.md                    # Homebrew setup guide
│   └── SETUP_SUMMARY.md             # This file!
├── scripts/
│   └── update-formula.sh            # Formula update automation
├── Makefile                         # Updated with build-all targets
├── README.md                        # Updated with Homebrew instructions
├── DISTRIBUTION.md                  # Complete distribution guide
└── RELEASE_CHECKLIST.md             # Release process checklist
```

### What Goes in the Tap Repository

```
homebrew-openscribe/
├── Formula/
│   └── openscribe.rb                # Your formula (copy from homebrew/openscribe.rb)
└── README.md                        # Brief description and install instructions
```

---

## 🎯 Quick Commands

```bash
# Build locally
make build-all

# Create release
git tag v0.1.0 && git push origin v0.1.0

# Update formula (after release is created)
./scripts/update-formula.sh 0.1.0 /path/to/homebrew-openscribe

# Test installation
brew tap alexandrelam/openscribe
brew install openscribe

# Users update
brew update
brew upgrade openscribe
```

---

## 📚 Documentation Reference

- **Quick Start:** See [DISTRIBUTION.md](../DISTRIBUTION.md) for complete step-by-step guide
- **Release Process:** See [RELEASE_CHECKLIST.md](../RELEASE_CHECKLIST.md) for checklist
- **Homebrew Details:** See [homebrew/README.md](README.md) for technical details
- **User Installation:** See [README.md](../README.md) for user-facing docs

---

## ✨ What You Get

Once set up, your users can:

1. **Easy Installation:**
   ```bash
   brew tap alexandrelam/openscribe
   brew install openscribe
   ```

2. **Automatic Updates:**
   ```bash
   brew update
   brew upgrade openscribe
   ```

3. **Dependency Management:**
   - whisper-cpp automatically installed
   - Proper permissions guidance
   - Post-install instructions

4. **Multi-Architecture Support:**
   - Automatic detection of Apple Silicon vs Intel
   - Optimized binaries for each platform

---

## 🔧 Maintenance

### Updating Dependencies

If you need to add/remove dependencies, edit the formula:

```ruby
depends_on "whisper-cpp"
depends_on "some-new-dependency"
```

### Updating Post-Install Messages

Edit the `caveats` section in `homebrew/openscribe.rb` to change the post-install instructions.

### Changing Installation

Modify the `install` block in the formula if you need to:
- Install additional files
- Create symlinks
- Set up configuration files

---

## 🆘 Getting Help

If you run into issues:

1. **Check the logs:**
   - GitHub Actions logs (for build failures)
   - `brew install --verbose openscribe` (for installation issues)

2. **Validate the formula:**
   ```bash
   brew audit --strict openscribe
   ```

3. **Review documentation:**
   - [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
   - [How to Create a Tap](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)

4. **Common issues:**
   - SHA256 mismatch: Re-download checksums from release
   - Build failures: Check GitHub Actions logs
   - Formula errors: Run `brew audit openscribe`

---

## ✅ Phase 13 Complete!

Everything is now ready for Homebrew distribution. Follow the "Next Steps" section above to:

1. Create the `homebrew-openscribe` repository
2. Make your first release (v0.1.0)
3. Update the formula
4. Test and announce!

Good luck! 🚀
