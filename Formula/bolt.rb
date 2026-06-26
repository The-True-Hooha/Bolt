class BoltFm < Formula
  desc "Bolt CLI tool"
  homepage "https://github.com/The-True-Hooha/Bolt"
  version "0.1.5"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/The-True-Hooha/Bolt/releases/download/v#{version}/bolt-fm_#{version}_darwin_arm64.tar.gz"
      # update after each release
      sha256 "245f19c97f71be9ce643972f766628b179f4a938c5dc9dccc09eb17e65c23a3a"

      def install
        bin.install "bolt-fm_#{version}_darwin_arm64" => "bolt-fm"
      end
    end

    on_intel do
      url "https://github.com/The-True-Hooha/Bolt/releases/download/v#{version}/bolt-fm_#{version}_darwin_amd64.tar.gz"
      # update after each release
      sha256 "724419a82b59a08169d384111ab3d1080900da4c775a3a276993022ba0a2706e"

      def install
        bin.install "bolt-fm_#{version}_darwin_amd64" => "bolt-fm"
      end
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/The-True-Hooha/Bolt/releases/download/v#{version}/bolt-fm_#{version}_linux_arm64.tar.gz"
      # update after each release
      sha256 "609da5de2814960625372affeaccf9bff6bd4fbf7185f819488f52f2e2cd0376"

      def install
        bin.install "bolt-fm_#{version}_linux_arm64" => "bolt-fm"
      end
    end

    on_intel do
      url "https://github.com/The-True-Hooha/Bolt/releases/download/v#{version}/bolt-fm_#{version}_linux_amd64.tar.gz"
      # update after each release
      sha256 "eb32c2cb58ee0bd5883ed8e17b7c3127b32bbe85de3501a5fdc39d0f2fd30f34"

      def install
        bin.install "bolt-fm_#{version}_linux_amd64" => "bolt-fm"
      end
    end
  end

  test do
    system "#{bin}/bolt-fm", "--version"
  end
end
