class N8r < Formula
  desc "The n8r runtime"
  homepage "https://github.com/injectionator/n8r"
  url "https://github.com/injectionator/n8r/releases/download/v1.0.0/n8r-1.0.0.tar.gz"
  sha256 "TODO"
  license "Copyright 2026 Steve Chambers, Injectionator, Viewyonder"

  def install
    bin.install "n8r"
  end

  test do
    system "#{bin}/n8r", "--version"
  end
end
