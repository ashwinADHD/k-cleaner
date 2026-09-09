#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
APP_NAME="K-Cleaner.app"
APP_DIR="$ROOT/$APP_NAME"
ICONSET="$ROOT/build/icon.iconset"

cd "$ROOT"

echo "Building kclean binary…"
go build -ldflags="-s -w" -o "$ROOT/kclean" .

echo "Creating app bundle…"
rm -rf "$APP_DIR"
mkdir -p "$APP_DIR/Contents/MacOS"
mkdir -p "$APP_DIR/Contents/Resources"

cp "$ROOT/Info.plist" "$APP_DIR/Contents/Info.plist"
cp "$ROOT/kclean" "$APP_DIR/Contents/MacOS/kclean"
chmod +x "$APP_DIR/Contents/MacOS/kclean"

mkdir -p "$ICONSET"
go run "$ROOT/scripts/genicon.go" "$ICONSET"

iconutil -c icns "$ICONSET" -o "$APP_DIR/Contents/Resources/AppIcon.icns"
/usr/libexec/PlistBuddy -c "Add :CFBundleIconFile string AppIcon" "$APP_DIR/Contents/Info.plist" 2>/dev/null || \
	/usr/libexec/PlistBuddy -c "Set :CFBundleIconFile AppIcon" "$APP_DIR/Contents/Info.plist"

echo ""
echo "Done: $APP_DIR"
echo ""
echo "To install:"
echo "  cp -R \"$APP_DIR\" /Applications/"
echo ""
echo "First launch (unsigned): Right-click K-Cleaner → Open"
