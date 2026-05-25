# MCP Server Manager

MCP サーバーの接続・ヘルスチェック・ツール一覧・テスト実行を一画面で管理するスタンドアロンデスクトップアプリ。

## 機能

- 各 MCP サーバーの状態（接続中 / 切断 / エラー / デグレード）をリアルタイム表示
- stdio（子プロセス）・HTTP トランスポート両対応
- Claude Desktop の `claude_desktop_config.json` からワンクリックでインポート
- 接続中サーバーのツール一覧を自動取得
- ツールの手動実行 — 入力 JSON を編集して即実行
- 30 秒ヘルスチェックループ＋サーキットブレーカー（3 連続失敗 → degraded）
- 接続ログのリアルタイム表示

## 技術スタック

| レイヤー | 技術 |
|----------|------|
| デスクトップ | [Wails v2](https://wails.io/) — Go をネイティブデスクトップ化 |
| バックエンド | Go 1.22+ |
| フロントエンド | Vue 3 + Vite + Pinia |
| スタイリング | UnoCSS + shadcn-vue |
| DB | SQLite + GORM（`~/.mcp-server-manager/`） |
| プロトコル | MCP JSON-RPC 2.0（stdio + HTTP） |

## 開発環境セットアップ

```bash
# 依存インストール
go mod download
cd frontend && npm install && cd ..

# 開発モード（ホットリロード）
wails dev

# 本番ビルド（.app / .exe）
wails build
```

## 使い方

### 1. サーバーを追加する

「+ Add Server」をクリック → Name・Transport・Command（または URL）を入力 → Add。

**stdio 例（Claude Desktop 互換形式）:**
```
Name:      filesystem
Transport: stdio
Command:   npx
Args:      ["-y", "@modelcontextprotocol/server-filesystem", "/Users/you/Desktop"]
```

**HTTP 例:**
```
Name:      my-http-server
Transport: HTTP
URL:       http://localhost:8080
```

### 2. Claude Desktop からインポート

「↓ Import Claude Desktop」をクリック — `~/Library/Application Support/Claude/claude_desktop_config.json` を自動検出してサーバーを一括登録します（既存エントリはスキップ）。

### 3. 接続してツールを試す

サーバー行の「Connect」→「Browse Tools」→ ツールを選択 → Arguments JSON を編集 → 「▶ Execute」。

### 4. イベントログを見る

サーバー詳細ページの「Events」タブでリアルタイムの接続ログ・エラー・stderr を確認できます。

## プロジェクト構造

```
mcp-server-manager/
├── main.go              # Wails エントリーポイント
├── app.go               # Go バインディング（Vue から呼ばれるメソッド）
├── internal/
│   ├── db/              # GORM モデル + SQLite 初期化
│   ├── mcp/             # JSON-RPC 2.0 transport（stdio / HTTP）+ Client
│   ├── health/          # ヘルスチェックループ + サーキットブレーカー
│   └── importer/        # Claude Desktop config インポーター
├── frontend/
│   └── src/
│       ├── pages/       # ServersView / ToolsView / EventsView
│       └── stores/      # Pinia（servers）
└── docs/
    ├── tech-selection.md
    └── development-plan.md
```

## テスト

```bash
# Go テスト
go test ./...

# フロントエンドテスト
cd frontend && npm test
```

---

Part of the [comet-taskAI](../comet-taskAI/concept.md) ecosystem.
