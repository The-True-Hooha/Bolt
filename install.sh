#!/usr/bin/env bash
set -e

REPO="https://github.com/The-True-Hooha/Bolt"
API="https://api.github.com/repos/The-True-Hooha/Bolt"
INSTALL_DIR="${BOLT_INSTALL_DIR:-$HOME/.local/bin}"
BINARY="bolt"
FROM_SOURCE=0

GREEN='\033[0;32m'; CYAN='\033[0;36m'; RED='\033[0;31m'; RESET='\033[0m'
ok()   { echo -e "${GREEN}✓${RESET} $*"; }
info() { echo -e "${CYAN}→${RESET} $*"; }
fail() { echo -e "${RED}✗${RESET} $*" >&2; exit 1; }

for arg in "$@"; do
    case "$arg" in
        --from-source) FROM_SOURCE=1 ;;
    esac
done

echo ""
echo "  ⚡ Bolt installer"
echo ""

mkdir -p "$INSTALL_DIR"

if [ "$FROM_SOURCE" = "1" ]; then
    command -v go >/dev/null 2>&1 || fail "Go not found. Install from https://go.dev/dl/ or omit --from-source to download prebuilt."
    info "Go $(go version | awk '{print $3}') found"

    if [ -f "go.mod" ] && grep -q "Bolt" go.mod 2>/dev/null; then
        info "building from current directory"
        BUILD_DIR="."
    else
        BUILD_DIR="$(mktemp -d)/bolt"
        info "cloning $REPO"
        git clone --depth 1 "$REPO" "$BUILD_DIR"
    fi

    info "building bolt..."
    (cd "$BUILD_DIR" && go build -ldflags="-s -w" -o "$BINARY" .)
    ok "build complete"
    mv "$BUILD_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
else
    command -v curl >/dev/null 2>&1 || command -v wget >/dev/null 2>&1 || fail "curl or wget required"

    info "fetching latest release info..."

    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"
    case "$ARCH" in
        x86_64)        ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) fail "unsupported arch: $ARCH" ;;
    esac

    if command -v curl >/dev/null 2>&1; then
        RELEASE_JSON="$(curl -fsSL "$API/releases/latest")"
    else
        RELEASE_JSON="$(wget -qO- "$API/releases/latest")"
    fi

    TAG="$(echo "$RELEASE_JSON" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
    URL="$(echo "$RELEASE_JSON" | grep "browser_download_url" | grep "${OS}_${ARCH}" | grep -v '\.tar\.gz\|\.zip' | head -1 | sed 's/.*"browser_download_url": *"\([^"]*\)".*/\1/')"

    [ -n "$URL" ] || fail "no prebuilt for ${OS}_${ARCH} in release $TAG — try --from-source"

    info "downloading bolt $TAG (${OS}_${ARCH})"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$URL" -o "$INSTALL_DIR/$BINARY"
    else
        wget -qO "$INSTALL_DIR/$BINARY" "$URL"
    fi
    ok "downloaded"
fi

chmod +x "$INSTALL_DIR/$BINARY"
ok "installed to $INSTALL_DIR/$BINARY"

add_to_path() {
    local RC="$1"
    local LINE="export PATH=\"\$HOME/.local/bin:\$PATH\""
    [ -f "$RC" ] || return 0
    grep -q "$INSTALL_DIR" "$RC" 2>/dev/null && return 0
    echo "" >> "$RC"
    echo "# added by bolt install" >> "$RC"
    echo "$LINE" >> "$RC"
    ok "added PATH to $RC"
}

if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    add_to_path "$HOME/.bashrc"
    add_to_path "$HOME/.zshrc"
    if [ -d "$HOME/.config/fish" ]; then
        FISH_CFG="$HOME/.config/fish/config.fish"
        if ! grep -q "$INSTALL_DIR" "$FISH_CFG" 2>/dev/null; then
            echo "set -gx PATH \"$INSTALL_DIR\" \$PATH" >> "$FISH_CFG"
            ok "added PATH to $FISH_CFG"
        fi
    fi
    echo ""
    echo "  Restart your terminal or run:"
    echo "    export PATH=\"$INSTALL_DIR:\$PATH\""
else
    ok "$INSTALL_DIR already in PATH"
fi

echo ""
ok "bolt installed! run: bolt --version"
echo ""
