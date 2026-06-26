class BoltFm < Formula
  desc "Bolt CLI tool"
  homepage "https://github.com/The-True-Hooha/Bolt"
  version "0.1.1"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/The-True-Hooha/Bolt/releases/download/v#{version}/bolt-fm_#{version}_darwin_arm64.tar.gz"
      # update after each release
      sha256 "REPLACE_WITH_SHA256_darwin_arm64"

      def install
        bin.install "bolt-fm_#{version}_darwin_arm64" => "bolt-fm"
      end
    end

    on_intel do
      url "https://github.com/The-True-Hooha/Bolt/releases/download/v#{version}/bolt-fm_#{version}_darwin_amd64.tar.gz"
      # update after each release
      sha256 "REPLACE_WITH_SHA256_darwin_amd64"

      def install
        bin.install "bolt-fm_#{version}_darwin_amd64" => "bolt-fm"
      end
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/The-True-Hooha/Bolt/releases/download/v#{version}/bolt-fm_#{version}_linux_arm64.tar.gz"
      # update after each release
      sha256 "REPLACE_WITH_SHA256_linux_arm64"

      def install
        bin.install "bolt-fm_#{version}_linux_arm64" => "bolt-fm"
      end
    end

    on_intel do
      url "https://github.com/The-True-Hooha/Bolt/releases/download/v#{version}/bolt-fm_#{version}_linux_amd64.tar.gz"
      # update after each release
      sha256 "REPLACE_WITH_SHA256_linux_amd64"

      def install
        bin.install "bolt-fm_#{version}_linux_amd64" => "bolt-fm"
      end
    end
  end

  test do
    system "#{bin}/bolt-fm", "--version"
  end
end
