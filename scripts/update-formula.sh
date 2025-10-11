#!/bin/bash

# Script to update Homebrew formula with new release information
# Usage: ./scripts/update-formula.sh <version> <tap-repo-path>

set -e

VERSION=$1
TAP_REPO=$2

if [ -z "$VERSION" ] || [ -z "$TAP_REPO" ]; then
    echo "Usage: $0 <version> <tap-repo-path>"
    echo "Example: $0 0.1.0 ../homebrew-openscribe"
    exit 1
fi

# Remove 'v' prefix if present
VERSION=${VERSION#v}

echo "Updating Homebrew formula for version $VERSION..."

# Download SHA256 checksums from GitHub release
echo "Fetching ARM64 SHA256..."
ARM64_SHA=$(curl -sL "https://github.com/alexandrelam/openscribe/releases/download/v${VERSION}/openscribe-darwin-arm64.tar.gz.sha256" | awk '{print $1}')

echo "Fetching AMD64 SHA256..."
AMD64_SHA=$(curl -sL "https://github.com/alexandrelam/openscribe/releases/download/v${VERSION}/openscribe-darwin-amd64.tar.gz.sha256" | awk '{print $1}')

if [ -z "$ARM64_SHA" ] || [ -z "$AMD64_SHA" ]; then
    echo "Error: Could not fetch SHA256 checksums from release"
    echo "Make sure the release v${VERSION} exists and has the checksum files"
    exit 1
fi

echo "ARM64 SHA256: $ARM64_SHA"
echo "AMD64 SHA256: $AMD64_SHA"

# Update the formula
FORMULA_PATH="$TAP_REPO/Formula/openscribe.rb"

if [ ! -f "$FORMULA_PATH" ]; then
    echo "Error: Formula not found at $FORMULA_PATH"
    exit 1
fi

# Create a temporary formula with updated values
cat > "$FORMULA_PATH" << EOF
class Openscribe < Formula
  desc "Real-time speech transcription CLI for macOS with hotkey activation"
  homepage "https://github.com/alexandrelam/openscribe"
  version "$VERSION"
  license "MIT"

  if Hardware::CPU.arm?
    url "https://github.com/alexandrelam/openscribe/releases/download/v#{version}/openscribe-darwin-arm64.tar.gz"
    sha256 "$ARM64_SHA"
  else
    url "https://github.com/alexandrelam/openscribe/releases/download/v#{version}/openscribe-darwin-amd64.tar.gz"
    sha256 "$AMD64_SHA"
  end

  depends_on "whisper-cpp"
  depends_on :macos

  def install
    if Hardware::CPU.arm?
      bin.install "openscribe-darwin-arm64" => "openscribe"
    else
      bin.install "openscribe-darwin-amd64" => "openscribe"
    end
  end

  def caveats
    <<~EOS
      OpenScribe has been installed successfully!

      Before using OpenScribe, you need to:

      1. Run the setup command to download Whisper models:
         $ openscribe setup

      2. Grant Accessibility permissions:
         - Open System Preferences > Security & Privacy > Privacy
         - Select "Accessibility" from the left sidebar
         - Click the lock icon to make changes
         - Add Terminal (or your terminal app) to the list
         - Enable the checkbox for Terminal

      3. Grant Microphone permissions:
         - Open System Preferences > Security & Privacy > Privacy
         - Select "Microphone" from the left sidebar
         - Click the lock icon to make changes
         - Add Terminal (or your terminal app) to the list
         - Enable the checkbox for Terminal

      4. Start using OpenScribe:
         $ openscribe start

      For more information, visit: https://github.com/alexandrelam/openscribe
    EOS
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/openscribe version")
  end
end
EOF

echo "Formula updated successfully!"
echo ""
echo "Next steps:"
echo "1. Review the changes: cat $FORMULA_PATH"
echo "2. Test the formula: brew install --build-from-source $TAP_REPO/Formula/openscribe.rb"
echo "3. Commit and push:"
echo "   cd $TAP_REPO"
echo "   git add Formula/openscribe.rb"
echo "   git commit -m 'Update openscribe to v$VERSION'"
echo "   git push origin main"
