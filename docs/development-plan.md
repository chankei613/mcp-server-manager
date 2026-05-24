# 開発計画

**作成日:** 2026-05-24  
**ターゲットリリース:** v0.1.0 (4週間)

---

## 開発ゴール — MVP v0.1.0

**コアコンセプト: "Postman for MCP"**

MCPを使い始めた開発者が、サーバーの接続状態を一目で確認し、ツールを手軽に試せるスタンドアロンツール。

### ユーザーストーリー（完了定義）

| # | ストーリー | 完了条件 |
|---|-----------|---------|
| 1 | MCPサーバーを追加・編集・削除できる | stdio / HTTP transport 両対応。設定はSQLiteに永続化 |
| 2 | 全サーバーの接続状態をリアルタイム確認できる | connected / disconnected / error / degraded をバッジで表示 |
| 3 | 接続中サーバーのツール一覧を自動取得できる | `tools/list` でツール名・説明・inputSchemaを表示 |
| 4 | 任意のツールをパラメータ付きで実行できる | inputSchemaからフォームを自動生成 → 実行 → JSON結果表示 |
| 5 | 接続イベント・エラーをログで確認できる | タイムスタンプ付きリアルタイムログ |
| ★ | Claude Desktop設定をワンクリックでインポート | `~/Library/Application Support/Claude/claude_desktop_config.json` 読み込み |

★ = Stretch goal（v0.1.0に含める）

### v0.1.0 スコープ外

- Resources API / Prompts API
- マルチワークスペース / プロファイル
- クラウド同期
- プラグインシステム
- ツール実行履歴の検索・エクスポート

---

## スプリント計画

### Week 1: Foundation ✅

**目標:** MCPサーバーに接続し、ツール一覧を取得できる状態

- [x] Wails v2 プロジェクト初期化
- [ ] GitHub Actions（test / lint / release）
- [x] SQLite スキーマ + GORM モデル
- [x] MCP stdio transport（子プロセスのstart/stop/restart）
- [x] JSON-RPC 2.0: `initialize` + `tools/list`
- [x] Wails bindings（GoメソッドをVueから呼べる状態）

### Week 2: Core Logic ✅

**目標:** ツール実行 + 自動ヘルスチェックが動く状態

- [x] `tools/call` 実装
- [x] HTTP transport 実装
- [x] ヘルスチェックループ（30秒間隔）
- [x] サーキットブレーカー（3連続失敗 → degraded → 自動再接続試行）
- [x] Claude Desktop config importer（~/.claude.json 優先対応含む）
- [x] Wailsイベント（Go → Vue リアルタイム通知）

### Week 3: UI ✅

**目標:** 全機能がUIから操作できる状態

- [x] アプリシェル（サイドバー + メインコンテンツ）
- [x] サーバーリストビュー（ステータスバッジ + CRUD）
- [x] ツール一覧ビュー（サーバー詳細ページ、スキーマベースフォーム）
- [x] ツール実行パネル（inputSchemaからフォーム自動生成 → 実行 → JSON結果）
- [x] イベントログビュー（リアルタイム）
- [x] インポートUI（Claude Desktop / Claude Code / Cursor / Windsurf / カスタム）

### Week 4: Polish + Distribution（進行中）

**目標:** v0.1.0リリース

- [x] エラー状態・空状態のハンドリング
- [x] 言語選択（日本語 / 英語 切り替え）
- [x] 起動時バージョンチェック（GitHub Releases API）
- [x] ランディングページ作成（MCPilot、Vercel デプロイ用）
- [x] 配布・販売・アップデート戦略ドキュメント（docs/distribution.md）
- [ ] Apple Developer 登録・コード署名・公証スクリプト整備
- [ ] Stripe Payment Link 設定・LP に反映
- [ ] LP をVercelにデプロイ
- [ ] スクリーンショット撮影・LP に反映
- [ ] README 使い方セクション + スクリーンショット
- [ ] v0.1.0 GitHub Release（DMG アップロード）
- [ ] ProductHunt 掲載
- [ ] 設定画面（スコープ外に移動）
- [ ] クロスプラットフォーム（Windows）対応（スコープ外に移動）

---

## ハーネス

### テスト

| 対象 | ツール | テスト範囲 |
|------|-------|---------|
| Go | `testing` + `testify` | JSON-RPC パース、サーキットブレーカー、DB操作、Claude Desktop importer |
| Vue | `vitest` + `@vue/test-utils` | Piniaストア、ユーティリティ関数 |
| E2E | スキップ（v0.1.0） | Wailsの自動E2EはMVP後 |

### Lint / Format

```
Go:   golangci-lint（gofmt + govet + staticcheck + errcheck）
JS:   eslint + eslint-plugin-vue + typescript-eslint
Fmt:  prettier（TS/Vue）
```

### CI/CD（GitHub Actions）

| Workflow | トリガー | 内容 |
|---------|---------|------|
| `test.yml` | PR作成・更新 | `go test ./...` + `vitest run` |
| `lint.yml` | PR作成・更新 | golangci-lint + eslint |
| `release.yml` | `v*.*.*` タグ push | Wailsマトリクスビルド → GitHub Release |

**リリースビルドマトリクス:**
- macOS arm64 (Apple Silicon)
- macOS amd64 (Intel)
- Windows amd64

### Git ワークフロー

```
ブランチ戦略:
  main         → stable / release only (PRでのみマージ)
  feat/xxx     → 機能追加
  fix/xxx      → バグ修正
  chore/xxx    → 設定・依存更新

コミット規約（Conventional Commits）:
  feat: add MCP stdio transport
  fix: circuit breaker reset timing
  chore: update wails to v2.10
  docs: update README with usage examples
```

### バージョン戦略

```
v0.1.0  stdio + HTTP transport、ツールブラウザ、実行パネル、Claude Desktop import
v0.2.0  Resources API、Prompts API、パフォーマンス改善
v1.0.0  安定版、全MCPメソッド対応、ドキュメント整備
```

---

## プロジェクト構造

```
mcp-server-manager/
├── .github/
│   └── workflows/
│       ├── test.yml
│       ├── lint.yml
│       └── release.yml
├── internal/
│   ├── mcp/         # MCP protocol (JSON-RPC 2.0, stdio/HTTP transport)
│   ├── db/          # GORM models + migration
│   ├── health/      # health check loop + circuit breaker
│   └── importer/    # Claude Desktop config import
├── frontend/        # Vue 3 + Vite (Wails標準配置)
│   └── src/
│       ├── components/
│       ├── pages/
│       └── stores/  # Pinia
├── docs/            # 設計ドキュメント
├── app.go           # Wails App struct + bindings
├── main.go          # エントリーポイント
├── wails.json       # Wails設定
├── go.mod
└── Makefile
```

---

## 思考・判断ログ

### 2026-05-24: Backend必要か問題

**問い:** MCP管理はローカル完結なのに別プロセスのbackendは必要か？

**判断:** Wailsを使うことでGoがデスクトップアプリの本体になる。HTTPサーバープロセスは不要。GoのメソッドをWails bindingsでVueから直接呼ぶ。MCP子プロセス管理（stdio）はGoのos/execで処理。

**理由:** MCP serverはstdioかHTTPで通信するが、そのプロセスを管理するにはネイティブコード（Go/Rust/Node）が必要。WailsはGoをデスクトップアプリとして直接使えるため、別途バックエンドプロセスが不要になる最もシンプルな構成。

### 2026-05-24: Nuxt → Vue 3 + Vite への変更

**問い:** フロントエンドはNuxt 3でよいか？

**判断:** Nuxt → Vue 3 + Vitaに変更。

**理由:** Nuxtのコアバリュー（SSR/SSG/ファイルベースルーティング）はデスクトップアプリでは不要かつ邪魔。WailsのVue 3テンプレートはVite直接使用。ビルドが軽く、Wailsとの相性が最良。

### 2026-05-24: UnoCSS採用

**問い:** Tailwind CSSの代替があるか？

**判断:** UnoCSS + shadcn-vue を採用。

**理由:** UnoCSS はTailwind互換プリセットを持ち、構文はほぼ同一。ビルド速度が5〜10倍速い。shadcn-vueはdev tool向けUIコンポーネント（テーブル、バッジ、ダイアログ）が揃っており、コピペ方式でカスタマイズ自由度が高い。
