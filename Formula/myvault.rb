class Myvault < Formula
  desc "Local-first, CLI-first secret manager for developers"
  homepage "https://github.com/rakshithgoud453/myvault"
  url "https://github.com/rakshithgoud453/myvault/archive/refs/tags/v0.3.0.tar.gz"
  sha256 "3ea1a323d976e5d44d7fa397c081bea3dc60a4aa732979999fa15b8a55008b95"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-ldflags", "-X github.com/rakshithgoud453/myvault/internal/cli.Version=#{version} -w -s", "-o", bin/"myvault", "./cmd/myvault"
  end

  test do
    assert_match "myvault version", shell_output("#{bin}/myvault version")
  end
end
