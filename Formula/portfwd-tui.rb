class PortfwdTui < Formula
  desc "TUI for kubectl port-forward across multiple targets"
  homepage "https://github.com/eljoe182/port-fordward-tui"
  version "1.3.0"

  on_macos do
    on_arm do
      url "https://github.com/eljoe182/port-fordward-tui/releases/download/v1.3.0/portfwd-tui-v1.3.0-darwin-arm64.tar.gz"
      sha256 "a194d564750d4124649ffdcb74f22652adcb8688124183207cc5c914702699c8"
    end
    on_intel do
      url "https://github.com/eljoe182/port-fordward-tui/releases/download/v1.3.0/portfwd-tui-v1.3.0-darwin-amd64.tar.gz"
      sha256 "0514d3d7cfcf9c27d19fbbdc7dbe64bf76d9460f4955a23f5652a798de894645"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/eljoe182/port-fordward-tui/releases/download/v1.3.0/portfwd-tui-v1.3.0-linux-arm64.tar.gz"
      sha256 "efe96da37effa20f140a517894dbe7416b5a77f755160fca04b8622ac96649ab"
    end
    on_intel do
      url "https://github.com/eljoe182/port-fordward-tui/releases/download/v1.3.0/portfwd-tui-v1.3.0-linux-amd64.tar.gz"
      sha256 "22859b259342d990e3c459d306ffe063e70df3b20bf61b91103f7101308f499c"
    end
  end

  def install
    bin.install "portfwd-tui"
  end

  test do
    assert_predicate bin/"portfwd-tui", :executable?
  end
end
