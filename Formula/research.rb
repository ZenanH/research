class Research < Formula
  desc "Fetch OpenAlex journal papers and export AI-friendly Markdown"
  homepage "https://github.com/ZenanH/research"
  url "https://github.com/ZenanH/research/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "REPLACE_WITH_RELEASE_SHA256"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-ldflags", "-s -w -X main.version=#{version}", "-o", bin/"research", "./cmd/research"
  end

  test do
    system "#{bin}/research", "--version"
  end
end

