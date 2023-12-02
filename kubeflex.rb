# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Kubeflex < Formula
  desc ""
  homepage "https://github.com/kubestellar/kubeflex"
  version "0.3.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/kubestellar/kubeflex/releases/download/v0.3.0/kubeflex_0.3.0_darwin_arm64.tar.gz"
      sha256 "0151ec82b4fecb73605196e482d2ee3844701c015ba4f715e950da997b4c8f04"

      def install
        bin.install "bin/kflex"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/kubestellar/kubeflex/releases/download/v0.3.0/kubeflex_0.3.0_darwin_amd64.tar.gz"
      sha256 "e3ffb4cee9fb74e9bceec250802be3693c3406ae851c0a043e6c79cec00a69a6"

      def install
        bin.install "bin/kflex"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/kubestellar/kubeflex/releases/download/v0.3.0/kubeflex_0.3.0_linux_arm64.tar.gz"
      sha256 "8d6954428778cdc0cabeebcf36cf8723ca3016f34a9dc048ddb204c06adf7735"

      def install
        bin.install "bin/kflex"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/kubestellar/kubeflex/releases/download/v0.3.0/kubeflex_0.3.0_linux_amd64.tar.gz"
      sha256 "fbbec24a73744cdf0b046f4ebd00e5167b870e65f9a2c0d2268626f15674b05b"

      def install
        bin.install "bin/kflex"
      end
    end
  end
end
