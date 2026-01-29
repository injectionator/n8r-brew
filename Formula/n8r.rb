class N8r < Formula
  desc "The n8r CLI - Injectionator command-line tool"
  homepage "https://injectionator.com"
  url "https://github.com/injectionator/n8r-brew/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "49ad7b3d8794392c3bb9d7c81031541fb02a12acc27ec9cbfe50b7b25f40e072"
  license "Copyright 2026 Steve Chambers, Injectionator, Viewyonder"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X github.com/injectionator/n8r/internal/config.Version=#{version}"), "./cmd/n8r"
  end

  test do
    system "#{bin}/n8r", "--version"
  end
end
