#!/usr/bin/env bash
set -e

REPO="https://github.com/The-True-Hooha/Bolt"
INSTALL_DIR="${BOLT_INSTALL_DIR:-$HOME/.local/bin}"
BINARY="bolt"

GREEN='\033[0;32m'; CYAN='\033[0;36m'; RED='\033[0;31m'; RESET='\033[0m'
ok()   { echo -e "${GREEN}✓${RESET} $*"; }
info() { echo -e "${CYAN}→${RESET} $*"; }
fail() { echo -e "${RED}✗${RESET} $*" >&2; exit 1; }

echo ""
echo "  ⚡ Bolt installer"
echo ""

command -v go >/dev/null 2>&1 || fail "Go is not installed. Get it from https://go.dev/dl/"
info "Go $(go version | awk '{print $3}') found"

if [ -f "go.mod" ] && grep -q "Bolt" go.mod 2>/dev/null; then
    info "building from current directory"
    BUILD_DIR="."
else
    BUILD_DIR="$(mktemp -d)/bolt"
    info "cloning $REPO"
    git clone --depth 1 "$REPO" "$BUILD_DIR"
fi

info "building bolt…"
(cd "$BUILD_DIR" && go build -ldflags="-s -w" -o "$BINARY" .)
ok "build complete"

mkdir -p "$INSTALL_DIR"
mv "$BUILD_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
chmod +x "$INSTALL_DIR/$BINARY"
ok "installed to $INSTALL_DIR/$BINARY"

add_to_path() {
    local RC="$1"
    local LINE='export PATH="$HOME/.local/bin:$PATH"'
    if [ -f "$RC" ] && grep -q "$INSTALL_DIR" "$RC" 2>/dev/null; then
        return 0
    fi
    if [ -f "$RC" ]; then
        echo "" >> "$RC"
        echo "# added by bolt install" >> "$RC"
        echo "$LINE" >> "$RC"
        ok "added PATH to $RC"
    fi
}

if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    add_to_path "$HOME/.bashrc"
    add_to_path "$HOME/.zshrc"
    if [ -d "$HOME/.config/fish" ]; then
        FISH_LINE="set -gx PATH \"$INSTALL_DIR\" \$PATH"
        FISH_CFG="$HOME/.config/fish/config.fish"
        if ! grep -q "$INSTALL_DIR" "$FISH_CFG" 2>/dev/null; then
            echo "$FISH_LINE" >> "$FISH_CFG"
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
