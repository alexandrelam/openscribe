class Openscribe < Formula
  desc "Real-time speech transcription CLI for macOS Apple Silicon with hotkey activation"
  homepage "https://github.com/alexandrelam/openscribe"
  version "0.1.0"
  license "MIT"

  url "https://github.com/alexandrelam/openscribe/releases/download/v#{version}/openscribe-darwin-arm64.tar.gz"
  sha256 "REPLACE_WITH_ARM64_SHA256"

  depends_on "whisper-cpp"
  depends_on :macos
  depends_on arch: :arm64

  def install
    bin.install "openscribe-darwin-arm64" => "openscribe"
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
