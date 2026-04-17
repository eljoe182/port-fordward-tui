class PortfwdTui < Formula
  desc "TUI for kubectl port-forward across multiple targets"
  homepage "https://github.com/eljoe182/port-fordward-tui"
  version "1.1.0"

  on_macos do
    on_arm do
      url "https://github.com/eljoe182/port-fordward-tui/releases/download/v1.1.0/portfwd-tui-v1.1.0-darwin-arm64.tar.gz"
      sha256 "c317e0b49d704161be19e4b5ee4e48f85d81ff84fb4025e7bde957a3dd959cfe"
    end
    on_intel do
      url "https://github.com/eljoe182/port-fordward-tui/releases/download/v1.1.0/portfwd-tui-v1.1.0-darwin-amd64.tar.gz"
      sha256 "119f03648011f3748bfe8c6625369617ecd8dbfaf24c39e78dc86968243f1fb9"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/eljoe182/port-fordward-tui/releases/download/v1.1.0/portfwd-tui-v1.1.0-linux-arm64.tar.gz"
      sha256 "163d2634946d3ce840b9c9fb6dcf93c287f7d55f209358bfdff8cc836a03a8cd"
    end
    on_intel do
      url "https://github.com/eljoe182/port-fordward-tui/releases/download/v1.1.0/portfwd-tui-v1.1.0-linux-amd64.tar.gz"
      sha256 "abd9a53a8ae5d82c04f6de00884a3381f7b231c252c24bee90a6ee0184c97f11"
    end
  end

  def install
    bin.install "portfwd-tui"
  end

  test do
    assert_predicate bin/"portfwd-tui", :executable?
  end
end
