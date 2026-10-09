#!/bin/sh
# Installs stremcli if it isn't already, then prints the binary's absolute path.
# Tries, in order: an existing install, `go install`, a prebuilt release binary.
set -eu

REPO="mebn/stremcli"
BIN_DIR="${STREMCLI_BIN_DIR:-$HOME/.local/bin}"

log() { echo "$@" >&2; }

find_existing() {
	if command -v stremcli >/dev/null 2>&1; then
		command -v stremcli
		return 0
	fi
	for dir in "$BIN_DIR" "${GOBIN:-}" "$(go env GOPATH 2>/dev/null)/bin" "$HOME/go/bin"; do
		if [ -n "$dir" ] && [ -x "$dir/stremcli" ]; then
			echo "$dir/stremcli"
			return 0
		fi
	done
	return 1
}

install_with_go() {
	command -v go >/dev/null 2>&1 || return 1
	log "Installing stremcli with go install..."
	go install "github.com/$REPO/cmd/stremcli@latest" >&2 || return 1
	gobin="$(go env GOBIN)"
	[ -n "$gobin" ] || gobin="$(go env GOPATH)/bin"
	echo "$gobin/stremcli"
}

install_release() {
	os="$(uname -s | tr '[:upper:]' '[:lower:]')"
	case "$(uname -m)" in
		x86_64 | amd64) arch=amd64 ;;
		arm64 | aarch64) arch=arm64 ;;
		*) log "Unsupported architecture: $(uname -m)"; return 1 ;;
	esac

	url="https://github.com/$REPO/releases/latest/download/stremcli-$os-$arch"
	log "Downloading $url..."
	mkdir -p "$BIN_DIR"
	tmp="$(mktemp)"
	if ! curl -fsSL "$url" -o "$tmp"; then
		rm -f "$tmp"
		return 1
	fi
	chmod +x "$tmp"
	mv "$tmp" "$BIN_DIR/stremcli"
	echo "$BIN_DIR/stremcli"
}

if path="$(find_existing)"; then
	echo "$path"
	exit 0
fi

if path="$(install_with_go)" || path="$(install_release)"; then
	log "Installed stremcli to $path"
	echo "$path"
	exit 0
fi

log "Could not install stremcli. Install Go (https://go.dev/dl) and run:"
log "  go install github.com/$REPO/cmd/stremcli@latest"
exit 1
