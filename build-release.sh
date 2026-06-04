#!/bin/bash
# MCPilot release build: wails build → codesign → notarize → staple → DMG
set -e

# ── 設定（要変更） ──────────────────────────────────────────────────────────
APPLE_ID="chankei613@gmail.com"
TEAM_ID="${TEAM_ID:-}"          # 環境変数 or 直書き（例: "ABC1234567"）
APP_PASSWORD="${APP_PASSWORD:-}" # App-specific password（環境変数推奨）

APP_NAME="MCPilot"
VERSION=$(grep 'AppVersion' version.go | sed 's/.*"\(.*\)".*/\1/')
APP_PATH="build/bin/${APP_NAME}.app"
ZIP_PATH="${APP_NAME}-${VERSION}.zip"
DMG_PATH="${APP_NAME}-${VERSION}.dmg"
ENTITLEMENTS="build/darwin/entitlements.plist"

# TEAM_ID が未設定なら Developer ID 一覧を表示して終了
if [ -z "$TEAM_ID" ]; then
  echo "ERROR: TEAM_ID が未設定です。"
  echo "以下のコマンドで Developer ID を確認してください:"
  echo "  security find-identity -v -p codesigning"
  echo ""
  echo "実行例:"
  echo "  TEAM_ID=ABC1234567 APP_PASSWORD=xxxx-xxxx-xxxx-xxxx ./build-release.sh"
  exit 1
fi

IDENTITY="Developer ID Application: アジサイ けいすけ (${TEAM_ID})"

echo "==> Building MCPilot v${VERSION}..."

# Dropbox の拡張属性を除去してからビルド
xattr -cr build/ 2>/dev/null || true
wails build -platform darwin/universal -o "${APP_NAME}"

echo "==> Code signing..."
codesign \
  --deep \
  --force \
  --verify \
  --verbose \
  --sign "${IDENTITY}" \
  --options runtime \
  --entitlements "${ENTITLEMENTS}" \
  "${APP_PATH}"

codesign --verify --deep --strict "${APP_PATH}"
echo "    Signature OK"

echo "==> Creating zip for notarization..."
ditto -c -k --keepParent "${APP_PATH}" "${ZIP_PATH}"

echo "==> Submitting for notarization (this takes a few minutes)..."
xcrun notarytool submit "${ZIP_PATH}" \
  --apple-id "${APPLE_ID}" \
  --team-id "${TEAM_ID}" \
  --password "${APP_PASSWORD}" \
  --wait

echo "==> Stapling notarization ticket..."
xcrun stapler staple "${APP_PATH}"
xcrun stapler validate "${APP_PATH}"
echo "    Staple OK"

echo "==> Creating DMG..."
rm -f "${DMG_PATH}"
hdiutil create \
  -volname "${APP_NAME}" \
  -srcfolder "${APP_PATH}" \
  -ov \
  -format UDZO \
  "${DMG_PATH}"

echo ""
echo "Done: ${DMG_PATH}"
echo ""
echo "次のステップ:"
echo "  1. GitHub Releases に ${DMG_PATH} をアップロード"
echo "  2. landing/thank-you.html の YOUR_DMG_DOWNLOAD_URL を更新"
echo "  3. npx vercel --prod で LP をデプロイ"
rm -f "${ZIP_PATH}"
