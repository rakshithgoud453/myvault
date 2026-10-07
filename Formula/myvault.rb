class Myvault < Formula
  desc "Local-first, CLI-first secret manager for developers"
  homepage "https://github.com/rakshithgoud453/myvault"
  url "https://github.com/rakshithgoud453/myvault/archive/refs/tags/v0.4.1.tar.gz"
  sha256 "f414f1f6aa107cb525d5587e4965f18540284c70f4a8a893f323320522eb569a"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-ldflags", "-X github.com/rakshithgoud453/myvault/internal/cli.Version=#{version} -w -s", "-o", bin/"myvault", "./cmd/myvault"
  end

  test do
    assert_match "myvault version", shell_output("#{bin}/myvault version")
  end
end
