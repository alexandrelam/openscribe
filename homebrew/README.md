# Homebrew Distribution for OpenScribe

This directory contains the Homebrew formula and setup instructions for distributing OpenScribe via Homebrew.

## Distribution Strategy

OpenScribe uses a **Homebrew Tap** approach for distribution. This means:
- You maintain your own Homebrew tap repository (separate from the main codebase)
- Users install OpenScribe via: `brew tap alexandrelam/openscribe && brew install openscribe`
- Updates are managed through your tap repository

## Setup Instructions

### 1. Create a Homebrew Tap Repository

Create a new GitHub repository named `homebrew-openscribe` (the `homebrew-` prefix is required by Homebrew convention):

```bash
# Create the repository on GitHub first, then:
git clone git@github.com:alexandrelam/homebrew-openscribe.git
cd homebrew-openscribe
```

### 2. Add the Formula

Copy the formula to your tap repository:

```bash
# From this directory
cp openscribe.rb /path/to/homebrew-openscribe/Formula/openscribe.rb
```

Or create the following structure in your tap repository:

```
homebrew-openscribe/
├── Formula/
│   └── openscribe.rb
└── README.md
```

### 3. Create a Release

Before publishing to Homebrew, you need to create a GitHub release:

```bash
# In your main openscribe repository
git tag v0.1.0
git push origin v0.1.0
```

This will trigger the GitHub Actions workflow (`.github/workflows/release.yml`) which:
- Builds binaries for both ARM64 and x86_64
- Creates tar.gz archives
- Generates SHA256 checksums
- Creates a GitHub release with all artifacts

### 4. Update the Formula with SHA256 Checksums

After the release is created, download the SHA256 files and update the formula:

```bash
# Get the checksums from the release
curl -L https://github.com/alexandrelam/openscribe/releases/download/v0.1.0/openscribe-darwin-arm64.tar.gz.sha256
curl -L https://github.com/alexandrelam/openscribe/releases/download/v0.1.0/openscribe-darwin-amd64.tar.gz.sha256
```

Update `openscribe.rb` with the actual SHA256 values:

```ruby
if Hardware::CPU.arm?
  url "https://github.com/alexandrelam/openscribe/releases/download/v0.1.0/openscribe-darwin-arm64.tar.gz"
  sha256 "actual_arm64_sha256_here"
else
  url "https://github.com/alexandrelam/openscribe/releases/download/v0.1.0/openscribe-darwin-amd64.tar.gz"
  sha256 "actual_amd64_sha256_here"
end
```

### 5. Commit and Push the Formula

```bash
cd homebrew-openscribe
git add Formula/openscribe.rb
git commit -m "Add openscribe formula v0.1.0"
git push origin main
```

### 6. Test the Installation

Test the formula locally before announcing:

```bash
# Tap your repository
brew tap alexandrelam/openscribe

# Install OpenScribe
brew install openscribe

# Verify installation
openscribe version
```

### 7. Announce Your Tap

Update your main repository README with installation instructions:

```markdown
## Installation

### Homebrew (recommended)

```bash
brew tap alexandrelam/openscribe
brew install openscribe
openscribe setup
```
```

## Updating the Formula for New Releases

When you release a new version:

1. Create a new tag in the main repository:
   ```bash
   git tag v0.2.0
   git push origin v0.2.0
   ```

2. Wait for the GitHub Actions workflow to complete and create the release

3. Get the new SHA256 checksums:
   ```bash
   curl -L https://github.com/alexandrelam/openscribe/releases/download/v0.2.0/openscribe-darwin-arm64.tar.gz.sha256
   curl -L https://github.com/alexandrelam/openscribe/releases/download/v0.2.0/openscribe-darwin-amd64.tar.gz.sha256
   ```

4. Update the formula in `homebrew-openscribe`:
   ```ruby
   version "0.2.0"
   # Update URLs and SHA256s
   ```

5. Commit and push:
   ```bash
   git add Formula/openscribe.rb
   git commit -m "Update openscribe to v0.2.0"
   git push origin main
   ```

6. Users can then update with:
   ```bash
   brew update
   brew upgrade openscribe
   ```

## Testing the Formula

### Local Testing

Test the formula locally before pushing:

```bash
# Audit the formula
brew audit --strict --online openscribe

# Test installation
brew install --build-from-source openscribe

# Test the binary
openscribe version
openscribe --help
```

### Testing on a Clean System

For thorough testing, use a clean macOS environment:

```bash
# Uninstall first
brew uninstall openscribe
brew untap alexandrelam/openscribe

# Then re-install
brew tap alexandrelam/openscribe
brew install openscribe
```

## Formula Maintenance

### Common Formula Updates

- **Version bump**: Update `version` field
- **SHA256 update**: Update checksums after each release
- **Dependency changes**: Update `depends_on` lines if needed
- **Installation path changes**: Modify `install` block
- **Post-install messages**: Update `caveats` section

### Homebrew Best Practices

1. **Use semantic versioning**: Follow semver (e.g., v1.0.0, v1.1.0, v2.0.0)
2. **Test before releasing**: Always test formula updates locally
3. **Keep caveats helpful**: Provide clear post-install instructions
4. **Maintain dependencies**: Keep dependency list minimal and accurate
5. **Document breaking changes**: Use caveats to warn about breaking changes

## Troubleshooting

### Formula Installation Fails

```bash
# Check formula syntax
brew audit openscribe

# Check installation logs
brew install openscribe --verbose

# Force reinstall
brew reinstall openscribe
```

### SHA256 Mismatch

This means the downloaded file doesn't match the expected checksum:

1. Verify the release artifacts on GitHub are correct
2. Re-download the SHA256 files
3. Update the formula with the correct checksums
4. Push the updated formula

### Dependency Issues

If whisper-cpp is not found:

```bash
# Verify whisper-cpp is installed
brew install whisper-cpp

# Reinstall openscribe
brew reinstall openscribe
```

## Resources

- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Homebrew Tap Documentation](https://docs.brew.sh/Taps)
- [Homebrew Best Practices](https://docs.brew.sh/Formula-Cookbook#homebrew-terminology)
- [Creating Taps](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
