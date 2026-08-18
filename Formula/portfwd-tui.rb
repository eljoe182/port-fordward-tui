class PortfwdTui < Formula
  desc "TUI for kubectl port-forward across multiple targets"
  homepage "https://github.com/eljoe182/port-fordward-tui"
  version "1.2.0"

  on_macos do
    on_arm do
      url "https://github.com/eljoe182/port-fordward-tui/releases/download/v1.2.0/portfwd-tui-v1.2.0-darwin-arm64.tar.gz"
      sha256 "cb2de9bedf4ae20ac01d3e5f18c50f12362185e7d9f1ad3d3172d7f89f8d3bc3"
    end
    on_intel do
      url "https://github.com/eljoe182/port-fordward-tui/releases/download/v1.2.0/portfwd-tui-v1.2.0-darwin-amd64.tar.gz"
      sha256 "37800f64b323f2201449f1d4d8e22aefe191352b914683c3dc1446cf6e6fd194"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/eljoe182/port-fordward-tui/releases/download/v1.2.0/portfwd-tui-v1.2.0-linux-arm64.tar.gz"
      sha256 "54e58bd5fc260fa19321f6ced98194ec4bdda83644f07088d53225eec78c5d11"
    end
    on_intel do
      url "https://github.com/eljoe182/port-fordward-tui/releases/download/v1.2.0/portfwd-tui-v1.2.0-linux-amd64.tar.gz"
      sha256 "4a57a41e1a85395aa2e3bea855e76626222cad8e77032e9cb1575ee62932dd1b"
    end
  end

  def install
    bin.install "portfwd-tui"
  end

  test do
    assert_predicate bin/"portfwd-tui", :executable?
  end
end
