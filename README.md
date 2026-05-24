# MCP Server Manager

**"Postman for MCP"** — MCPサーバーの接続・ヘルスチェック・ツール一覧・テスト実行を一画面で管理するスタンドアロンツール。

## 概要

- 各MCPサーバーの状態（接続中/切断/エラー）をリアルタイム表示
- 接続しているサーバーのツール一覧を自動取得
- ツールの試し打ち（Postmanのような手動実行UI）
- stdio / HTTP+SSE トランスポート対応

## 技術スタック

- **Frontend**: Nuxt 3 + Vue 3 + Pinia + Tailwind CSS
- **Backend**: Go 1.22+ (chi router + GORM)
- **DB**: SQLite（ローカルファースト）
- **Protocol**: MCP JSON-RPC 2.0 (stdio + HTTP transport)
- **Real-time**: SSE (Server-Sent Events)

## 起動方法

```bash
# 開発
make dev

# ビルド (シングルバイナリ)
make build
./mcp-server-manager
```

## プロジェクト構造

```
mcp-server-manager/
├── backend/          # Go バックエンド
│   ├── cmd/server/   # エントリーポイント
│   └── internal/
│       ├── mcp/      # MCP プロトコル実装
│       ├── api/      # HTTP ハンドラー
│       ├── db/       # SQLite モデル
│       └── health/   # ヘルスチェックループ
├── frontend/         # Nuxt 3 フロントエンド
└── Makefile
```

---

Part of the [comet-taskAI](../comet-taskAI/concept.md) ecosystem.
