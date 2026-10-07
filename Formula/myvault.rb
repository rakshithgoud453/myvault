class Myvault < Formula
  desc "Local-first, CLI-first secret manager for developers"
  homepage "https://github.com/rakshithgoud453/myvault"
  url "https://github.com/rakshithgoud453/myvault/archive/refs/tags/v0.4.0.tar.gz"
  sha256 "446d23111bfc2fc58c91545c3697163bd0acdbe6aff09423ca3ddce1cb4b151e"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-ldflags", "-X github.com/rakshithgoud453/myvault/internal/cli.Version=#{version} -w -s", "-o", bin/"myvault", "./cmd/myvault"
  end

  test do
    assert_match "myvault version", shell_output("#{bin}/myvault version")
  end
end
