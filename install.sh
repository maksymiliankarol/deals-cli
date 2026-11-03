#!/bin/sh
set -eu

repo=maksymiliankarol/deals-cli
os=$(uname -s | tr '[:upper:]' '[:lower:]')

case "$os" in
  linux|darwin) ;;
  *) echo "unsupported OS: $os" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64)  arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

dir="${INSTALL_DIR:-$HOME/.local/bin}"
mkdir -p "$dir"

curl -fsSL "https://github.com/$repo/releases/latest/download/deals_${os}_${arch}.tar.gz" \
  | tar xz -C "$dir" deals

echo "Installed $dir/deals"
case ":$PATH:" in
  *":$dir:"*) ;;
  *) echo "Add $dir to your PATH to use it." ;;
esac