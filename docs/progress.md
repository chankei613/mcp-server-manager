# 進捗ログ

---

## 2026-05-24

### Phase 1–3 完了（Week 1〜3 相当）

- ✅ Wails v2 プロジェクト初期化・Go/Vue 基盤構築
- ✅ SQLite + GORM によるサーバー・イベント管理
- ✅ MCP stdio / HTTP transport 実装（JSON-RPC 2.0）
- ✅ ヘルスチェックループ・サーキットブレーカー
- ✅ サーバー一覧・接続トグル・削除 UI
- ✅ ツール詳細ページ（スキーマベースフォーム自動生成・実行・JSON結果表示）
- ✅ インポートUI（Claude Desktop / Claude Code / Cursor / Windsurf / カスタム JSON）
- ✅ 再起動時の stale 接続状態リセット（startup で DB をクリア）

**気づき:** Wailsビルドが Dropbox の xattr 問題で codesign 失敗する。`xattr -cr build/` を毎回実行する必要あり。`build-mac.sh` に組み込み済み。

---

## 2026-05-25

### Phase 4 — UX改善・env対応・i18n・配布戦略

- ✅ env セキュリティ実装：APIキーをDBに保存せず、Connect時に元設定ファイルから読み取り
  - `FindServerEnv()` が全既知設定ファイルをサーバー名で検索
  - `StdioTransport` に `env []string` フィールド追加、`os.Environ()` にマージして渡す
  - YouTube など env 必須サーバーの接続が通るようになった
- ✅ `~/.claude.json` 優先インポート対応（従来は `~/.claude/claude_desktop_config.json` のみ）
  - Claude Code は `~/.claude.json` にサーバー設定を持つ（11件）
  - `GetLocalClaudeConfigPath()` で候補ファイルを順番に検索
- ✅ ツールUI全面改善
  - スキーマベースフォーム（型別入力、説明文、required バッジ）
  - フォーム ⇔ JSON 切り替えトグル
  - 大きな「▶ 実行する」ボタン
  - サーバー情報パネル（コマンド全文、エラーログ）
- ✅ 言語選択機能（日本語 / 英語）
  - `stores/i18n.ts`：Pinia ストア + 翻訳テーブル（約80キー × 2言語）
  - サイドバー下部のトグルボタンで即時切り替え
  - `localStorage` に保存、再起動後も維持
  - ツール説明・パラメータ名はMCPサーバー側コンテンツのため翻訳しない方針
- ✅ 製品名を **MCPilot** に決定（"Postman for MCP" → "Visual Client for MCP Servers"）
  - Postman は登録商標のため LP での使用を避ける
- ✅ ランディングページ作成（`landing/index.html`）
  - 純粋な HTML + CSS + Vanilla JS（依存ゼロ、Vercel デプロイ対応）
  - ProductHunt スタイルの英語 LP
  - 日英切り替えボタン付き（`data-i18n` 属性 + 翻訳オブジェクト）
  - $9 買い切りの価格カード・機能リスト・3ステップ説明
- ✅ 起動時バージョンチェック実装
  - `version.go`：`AppVersion` 定数（一元管理）
  - `updater.go`：GitHub Releases API 呼び出し + semver 比較
  - `app.go`：`startup()` からバックグラウンドで実行
  - `App.vue`：`update:available` イベントを受けてインディゴ色バナー表示
- ✅ 配布・販売戦略ドキュメント作成（`docs/distribution.md`）
  - App Store 不採用理由（stdio プロセス起動のサンドボックル問題）
  - Apple Developer $99 / コード署名・公証手順
  - Stripe Payment Links による販売フロー
  - GitHub Releases + Vercel による配布構成
  - アップデートリリース手順

**使用Worker:** claude-sonnet-4-6

**気づき・改善点:**
- 言語選択は「UIはアプリ言語で、ツール内容（MCP サーバー由来）は翻訳しない」という線引きが重要
- LP の Hero に GitHub リンクを置くと購入 CTA と競合する → フッター・ナビのみに限定
- Stripe 単体でも国際販売は可能。Lemon Squeezy の優位性は Merchant of Record としての税務代行のみ

---

## 残タスク（v0.1.0 リリースまで）

- [ ] Apple Developer 登録・コード署名・公証スクリプト整備
- [ ] Stripe Payment Link 作成 → LP に反映
- [ ] Vercel デプロイ
- [ ] スクリーンショット撮影 → LP に反映
- [ ] README 使い方セクション更新
- [ ] v0.1.0 GitHub Release（署名済み DMG アップロード）
- [ ] ProductHunt 掲載
