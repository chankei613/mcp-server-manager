# MCPilot — 配布・販売・アップデート戦略

## 製品情報

| 項目 | 内容 |
|------|------|
| 製品名 | MCPilot |
| サブタイトル | Visual Client for MCP Servers |
| 価格 | $9（買い切り） |
| 対象 | macOS（開発者向け） |
| 配布方法 | 直接配布（App Store 外） |

---

## 配布方法の判断

### App Store を使わない理由

このアプリは `npx`・`uvx` などのローカルプロセスを子プロセスとして起動する。
App Store のサンドボックス制限によりこの動作が不可能になるため、**直接配布一択**。

### 配布構成

```
LP（Vercel）
  ↓ 購入ボタン
Stripe 決済
  ↓ 完了後リダイレクト
thank-you ページ
  ↓ ダウンロードリンク
GitHub Releases（DMG）
```

---

## コード署名・公証（必須）

署名・公証なしでは macOS が「開発元を検証できません」でブロックする。

### 必要なもの

- Apple Developer Program 登録（$99/年）
- Developer ID Application 証明書（Xcode から発行）

### ビルド〜配布の手順

```bash
# 1. ビルド（Dropbox の xattr 問題を先に解消）
xattr -cr build/
wails build

# 2. コード署名
codesign --deep --force --verify --verbose \
  --sign "Developer ID Application: YOUR NAME (TEAM_ID)" \
  --options runtime \
  build/bin/MCPilot.app

# 3. zip 化（公証はzipまたはdmgで送付）
ditto -c -k --keepParent build/bin/MCPilot.app MCPilot.zip

# 4. Apple に公証送付
xcrun notarytool submit MCPilot.zip \
  --apple-id "your@email.com" \
  --team-id "TEAM_ID" \
  --password "app-specific-password" \
  --wait

# 5. チケットを結合
xcrun stapler staple build/bin/MCPilot.app

# 6. DMG 化
hdiutil create -volname MCPilot -srcfolder build/bin/MCPilot.app \
  -ov -format UDZO MCPilot.dmg
```

> **Dropbox xattr 問題**：Dropbox がファイルに拡張属性を付加するため、
> `codesign` が失敗する。`xattr -cr build/` を毎回実行すること。

---

## 販売（Stripe）

### 設定手順

1. Stripe ダッシュボード → Products → Add product
   - 名前: MCPilot
   - 価格: $9 USD、一回限り（buy once）
2. Payment Links → Create link
   - 商品を選択
   - After payment: "Redirect to URL" → `https://mcpilot.dev/thank-you`（ドメイン確定後に設定）
3. LP の `YOUR_STRIPE_PAYMENT_LINK` を発行されたURLに差し替え

### ライセンスキーについて

現時点では実装しない。理由：

- $9 のツールに認証コストが見合わない
- 対象ユーザー（開発者）は誠実な人が多い
- Stripe の購入メールがレシート代わりになる

ユーザーが増えたタイミングで Lemon Squeezy への移行を検討する（Merchant of Record として税務対応も自動化できる）。

---

## ランディングページ

| 項目 | 内容 |
|------|------|
| ファイル | `landing/index.html` |
| ホスティング | Vercel |
| 技術 | 純粋な HTML + CSS + Vanilla JS（依存ゼロ） |
| 日英切り替え | `JA` / `EN` ボタン、`localStorage` に保存 |

### Vercel デプロイ手順

```bash
# Vercel CLI を使う場合
cd landing
npx vercel --prod

# または Vercel ダッシュボードから
# → New Project → Import Git Repository
# → Root Directory: landing
# → Framework: Other
```

### LP 差し替え作業（リリース前）

- [ ] `YOUR_STRIPE_PAYMENT_LINK` → Stripe の Payment Link URL
- [ ] DMG のダウンロード URL → GitHub Releases の最新 DMG URL
- [ ] スクリーンショット（UIモックアップを実際の画像に差し替え）

---

## アップデート配布

### 仕組み

起動時にバックグラウンドで GitHub Releases API を叩き、最新バージョンと比較する。
新しいバージョンがあれば `update:available` イベントを Vue に送信し、バナーを表示する。

```
起動
 ↓
api.github.com/repos/chankei613/mcp-server-manager/releases/latest
 ↓
tag_name の semver を比較
 ↓ 新しければ
Vue にイベント送信 → 画面上部にバナー表示
```

### 関連ファイル

| ファイル | 役割 |
|---------|------|
| `version.go` | `AppVersion` 定数を一元管理 |
| `updater.go` | GitHub API 呼び出しと semver 比較 |
| `app.go` | `startup()` からバックグラウンド起動 |
| `frontend/src/App.vue` | バナー表示 |

### 新バージョンのリリース手順

```
1. version.go の AppVersion を上げる（例: "0.1.0" → "0.2.0"）
2. wails build → 署名・公証 → DMG 作成
3. GitHub で新しいリリースを作成
   - タグ: v0.2.0
   - DMG をアセットとしてアップロード
4. 既存ユーザーが次回起動時にバナーで通知を受け取る
```

---

## ProductHunt 掲載について

- OSS・有料どちらでも掲載可能
- 開発者向けツールは OSS の方が注目されやすい傾向がある
- 署名なしでも掲載自体は問題ない（インストール方法の説明をREADMEに記載する）
- 掲載タイミング：LP・Stripe・署名が揃ってから

---

## 今後の検討事項

| 項目 | 優先度 | 条件 |
|------|--------|------|
| Sparkle 自動更新 | 低 | ユーザーが増えてから |
| ライセンスキー管理 | 低 | 月収が安定してから |
| Lemon Squeezy 移行 | 低 | 海外販売が本格化したら |
| Windows 対応 | 中 | 要望が出てから |
