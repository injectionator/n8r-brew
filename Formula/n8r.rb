class N8r < Formula
  desc "The n8r CLI - Injectionator command-line tool"
  homepage "https://injectionator.com"
  url "https://github.com/injectionator/homebrew-n8r-brew/releases/download/v0.2.0/n8r-0.2.0.tar.gz"
  sha256 "78ac6f30fe3a2afe369646d4af65b1b6ace46d4d589dfba190dff6c2ebe75960"
  license "Copyright 2026 Steve Chambers, Injectionator, Viewyonder"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X github.com/injectionator/n8r/internal/config.Version=#{version}"), "./cmd/n8r"
  end

  test do
    system "#{bin}/n8r", "--version"
  end
end
