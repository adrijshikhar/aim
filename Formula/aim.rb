# typed: false
# frozen_string_literal: true

class Aim < Formula
  desc "Isolated Profile Manager for AI Agents (Antigravity, Claude Code, Codex, Gemini)"
  homepage "https://github.com/adrijshikhar/aim"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/adrijshikhar/aim/releases/download/v#{version}/aim_#{version}_darwin_arm64.tar.gz"
    else
      url "https://github.com/adrijshikhar/aim/releases/download/v#{version}/aim_#{version}_darwin_amd64.tar.gz"
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/adrijshikhar/aim/releases/download/v#{version}/aim_#{version}_linux_arm64.tar.gz"
    else
      url "https://github.com/adrijshikhar/aim/releases/download/v#{version}/aim_#{version}_linux_amd64.tar.gz"
    end
  end

  def install
    bin.install "aim"

    # Install shell completions
    generate_completions_from_executable(bin/"aim", "completion")
  end

  test do
    assert_match "aim version #{version}", shell_output("#{bin}/aim --version")
  end
end
