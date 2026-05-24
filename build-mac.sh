#!/bin/bash
set -e
wails build
xattr -cr build/bin/mcp-server-manager.app
echo "✓ build/bin/mcp-server-manager.app を開けます"
