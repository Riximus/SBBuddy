#!/usr/bin/env bash
set -e
BIN_NAME=SBBuddy
DEST=${DEST:-/usr/local/bin}

echo "Installing $BIN_NAME to $DEST…"
install -m 0755 "$BIN_NAME" "$DEST/$BIN_NAME"
echo "Done! You can now run '$BIN_NAME' from anywhere."
