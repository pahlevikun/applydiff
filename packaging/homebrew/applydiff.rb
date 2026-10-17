# Template for a Homebrew tap (e.g. github.com/pahlevikun/homebrew-tap).
# Copy to Formula/applydiff.rb and replace VERSION / sha256 values after each release.
class Applydiff < Formula
  desc "Git diff for your CV — local-first job search, tailoring, and interview prep"
  homepage "https://github.com/pahlevikun/applydiff"
  version "VERSION"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/pahlevikun/applydiff/releases/download/vVERSION/applydiff_VERSION_darwin_arm64.tar.gz"
      sha256 "DARWIN_ARM64_SHA256"
    end
    on_intel do
      url "https://github.com/pahlevikun/applydiff/releases/download/vVERSION/applydiff_VERSION_darwin_amd64.tar.gz"
      sha256 "DARWIN_AMD64_SHA256"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/pahlevikun/applydiff/releases/download/vVERSION/applydiff_VERSION_linux_arm64.tar.gz"
      sha256 "LINUX_ARM64_SHA256"
    end
    on_intel do
      url "https://github.com/pahlevikun/applydiff/releases/download/vVERSION/applydiff_VERSION_linux_amd64.tar.gz"
      sha256 "LINUX_AMD64_SHA256"
    end
  end

  def install
    bin.install "applydiff"
  end

  def caveats
    <<~EOS
      PDF export needs a TeX engine. Optional: brew install tectonic
      (or use xelatex if it is already on your PATH).
    EOS
  end

  test do
    assert_match "ApplyDiff", shell_output("#{bin}/applydiff --help")
  end
end
