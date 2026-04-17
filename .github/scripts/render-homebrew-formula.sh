#!/usr/bin/env bash
# Renders Formula/portfwd-tui.rb for the Homebrew tap layout in this repository.
# Expects: GITHUB_REPOSITORY, GITHUB_REF_NAME; first arg: path to dist/ (contains SHA256SUMS + tarballs).
set -euo pipefail

DIST_DIR="${1:?usage: render-homebrew-formula.sh <dist-dir>}"
REPO="${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is required}"
TAG="${GITHUB_REF_NAME:?GITHUB_REF_NAME is required}"
ver="${TAG#v}"

get_sha() {
  local pair="$1"
  local fname="portfwd-tui-${TAG}-${pair}.tar.gz"
  awk -v f="$fname" '$NF == f { print $1; exit }' "${DIST_DIR}/SHA256SUMS"
}

SHA_DARWIN_ARM64=$(get_sha darwin-arm64)
SHA_DARWIN_AMD64=$(get_sha darwin-amd64)
SHA_LINUX_ARM64=$(get_sha linux-arm64)
SHA_LINUX_AMD64=$(get_sha linux-amd64)

check_sha() {
  local name="$1"
  local val="$2"
  if [[ -z "$val" ]]; then
    echo "missing sha256 for ${name}" >&2
    exit 1
  fi
}
check_sha darwin-arm64 "$SHA_DARWIN_ARM64"
check_sha darwin-amd64 "$SHA_DARWIN_AMD64"
check_sha linux-arm64 "$SHA_LINUX_ARM64"
check_sha linux-amd64 "$SHA_LINUX_AMD64"

cat <<EOF
class PortfwdTui < Formula
  desc "TUI for kubectl port-forward across multiple targets"
  homepage "https://github.com/${REPO}"
  version "${ver}"

  on_macos do
    on_arm do
      url "https://github.com/${REPO}/releases/download/${TAG}/portfwd-tui-${TAG}-darwin-arm64.tar.gz"
      sha256 "${SHA_DARWIN_ARM64}"
    end
    on_intel do
      url "https://github.com/${REPO}/releases/download/${TAG}/portfwd-tui-${TAG}-darwin-amd64.tar.gz"
      sha256 "${SHA_DARWIN_AMD64}"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/${REPO}/releases/download/${TAG}/portfwd-tui-${TAG}-linux-arm64.tar.gz"
      sha256 "${SHA_LINUX_ARM64}"
    end
    on_intel do
      url "https://github.com/${REPO}/releases/download/${TAG}/portfwd-tui-${TAG}-linux-amd64.tar.gz"
      sha256 "${SHA_LINUX_AMD64}"
    end
  end

  def install
    bin.install "portfwd-tui"
  end

  test do
    assert_predicate bin/"portfwd-tui", :executable?
  end
end
EOF
