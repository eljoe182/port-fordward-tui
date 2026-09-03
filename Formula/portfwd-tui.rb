class PortfwdTui < Formula
  desc "TUI for kubectl port-forward across multiple targets"
  homepage "https://github.com/eljoe182/port-fordward-tui"
  version "1.4.0"

  on_macos do
    on_arm do
      url "https://github.com/eljoe182/port-fordward-tui/releases/download/v1.4.0/portfwd-tui-v1.4.0-darwin-arm64.tar.gz"
      sha256 "1bb1345df1491301c378554faecad217a94fe96e180c63fb13a6c3f8f28725fc"
    end
    on_intel do
      url "https://github.com/eljoe182/port-fordward-tui/releases/download/v1.4.0/portfwd-tui-v1.4.0-darwin-amd64.tar.gz"
      sha256 "e8fb3845b5c24903001cef2cf4d03880eb7b24492dfadb84a6dd9929f5ab29ec"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/eljoe182/port-fordward-tui/releases/download/v1.4.0/portfwd-tui-v1.4.0-linux-arm64.tar.gz"
      sha256 "54c670d015a0486972f65587911692077442ab4b2a7a65cd2c57d4deee7f8f50"
    end
    on_intel do
      url "https://github.com/eljoe182/port-fordward-tui/releases/download/v1.4.0/portfwd-tui-v1.4.0-linux-amd64.tar.gz"
      sha256 "37eb7c8f5e004dea4f48a783010a5a6130d3dde94d83f5096a21e353b4060eb1"
    end
  end

  def install
    bin.install "portfwd-tui"
  end

  test do
    assert_predicate bin/"portfwd-tui", :executable?
  end
end
