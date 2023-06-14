# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Kubeflex < Formula
  desc ""
  homepage "https://github.com/kubestellar/kubeflex"
  version "0.1.0"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/kubestellar/kubeflex/releases/download/v0.1.0/kubeflex_0.1.0_darwin_amd64.tar.gz"
      sha256 "40f4b5dd6111add03161b46ef299d708bcad80a0b9169a2eba5bb87d4c365ed7"

      def install
        bin.install "bin/kflex"
      end
    end
    if Hardware::CPU.arm?
      url "https://github.com/kubestellar/kubeflex/releases/download/v0.1.0/kubeflex_0.1.0_darwin_arm64.tar.gz"
      sha256 "864d14614287de6449401f657511949b5244edc0807fc2baf931f49a09b3c2c5"

      def install
        bin.install "bin/kflex"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/kubestellar/kubeflex/releases/download/v0.1.0/kubeflex_0.1.0_linux_arm64.tar.gz"
      sha256 "f771f03b2eb7c58ec358d9f9186cfda65b0a49b7dd5d4739216536a7cfc9f169"

      def install
        bin.install "bin/kflex"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/kubestellar/kubeflex/releases/download/v0.1.0/kubeflex_0.1.0_linux_amd64.tar.gz"
      sha256 "7b2d13ca68f309f5b9032fe5b5d8beedf64c7489dfa98d959deb256ed97d2e56"

      def install
        bin.install "bin/kflex"
      end
    end
  end
end
