#!/usr/bin/env bash
set -euo pipefail

BINARY_NAME="wrep"
DEFAULT_INSTALL_DIR="${HOME}/.local/bin"

usage() {
    cat <<EOF
Usage: ./install.sh [OPTIONS]

Options:
  -d, --dir <DIR>    Install directory (default: ${DEFAULT_INSTALL_DIR})
  -h, --help         Show this help message

Environment variables:
  PREFIX             Install prefix (default: ~/.local)
                     Lower priority than --dir
EOF
}

INSTALL_DIR=""
while [[ $# -gt 0 ]]; do
    case "$1" in
        -d|--dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            usage >&2
            exit 1
            ;;
    esac
done

if [[ -z "${INSTALL_DIR}" ]]; then
    if [[ -n "${PREFIX:-}" ]]; then
        INSTALL_DIR="${PREFIX}/bin"
    else
        INSTALL_DIR="${DEFAULT_INSTALL_DIR}"
    fi
fi

if ! command -v go &>/dev/null; then
    echo "Error: go not found. Please install Go first." >&2
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}"

echo "Building: ${BINARY_NAME} ..."
go build -o "${BINARY_NAME}" .

mkdir -p "${INSTALL_DIR}"
install -m 0755 "${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"

echo "Installed: ${INSTALL_DIR}/${BINARY_NAME}"

if [[ ":${PATH}:" != *":${INSTALL_DIR}:"* ]]; then
    echo "Note: ${INSTALL_DIR} is not in your PATH." >&2
    echo "  Add the following line to ~/.bashrc or ~/.zshrc:" >&2
    echo "    export PATH=\"${INSTALL_DIR}:\$PATH\"" >&2
fi
