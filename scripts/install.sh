#!/usr/bin/env bash
# Install the latest applydiff release binary for this OS/arch.
set -euo pipefail

REPO="${APPLYDIFF_INSTALL_REPO:-pahlevikun/applydiff}"
INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "${os}" in
  linux) ;;
  darwin) ;;
  *)
    echo "install.sh: unsupported OS: ${os}" >&2
    exit 1
    ;;
esac

arch="$(uname -m)"
case "${arch}" in
  x86_64 | amd64) arch="amd64" ;;
  aarch64 | arm64) arch="arm64" ;;
  *)
    echo "install.sh: unsupported CPU: ${arch}" >&2
    exit 1
    ;;
esac

if [[ -n "${APPLYDIFF_VERSION:-}" ]]; then
  version="${APPLYDIFF_VERSION#v}"
  tag="v${version}"
else
  tag="$(
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" |
      python3 -c 'import json,sys; print(json.load(sys.stdin)["tag_name"])'
  )"
  version="${tag#v}"
fi

archive="applydiff_${version}_${os}_${arch}.tar.gz"
base="https://github.com/${REPO}/releases/download/${tag}"
tmpdir="$(mktemp -d)"
trap 'rm -rf "${tmpdir}"' EXIT

curl -fsSL "${base}/${archive}" -o "${tmpdir}/${archive}"
curl -fsSL "${base}/checksums.txt" -o "${tmpdir}/checksums.txt"

(
  cd "${tmpdir}"
  line="$(grep " ${archive}\$" checksums.txt || true)"
  if [[ -z "${line}" ]]; then
    echo "install.sh: ${archive} not found in checksums.txt" >&2
    exit 1
  fi
  echo "${line}" | sha256sum -c -
)

tar -xzf "${tmpdir}/${archive}" -C "${tmpdir}"
mkdir -p "${INSTALL_DIR}"
install -m 755 "${tmpdir}/applydiff" "${INSTALL_DIR}/applydiff"

echo "Installed applydiff ${version} -> ${INSTALL_DIR}/applydiff"
if ! command -v applydiff >/dev/null 2>&1; then
  echo "Add ${INSTALL_DIR} to your PATH if needed."
fi
