class Myvault < Formula
  desc "Local-first, CLI-first secret manager for developers"
  homepage "https://github.com/rakshithgoud453/myvault"
  url "https://github.com/rakshithgoud453/myvault/archive/refs/tags/v0.4.2.tar.gz"
  sha256 "c1de7941dc30074a8396cdbd57c14d35efb6c631e1f7f3919a6ae9d9a8df4338"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-ldflags", "-X github.com/rakshithgoud453/myvault/internal/cli.Version=#{version} -w -s", "-o", bin/"myvault", "./cmd/myvault"
  end

  test do
    assert_match "myvault version", shell_output("#{bin}/myvault version")
  end
end
