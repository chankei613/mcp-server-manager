# 技術選定記録

**作成日:** 2026-05-24  
**ステータス:** 確定

---

## 採用スタック

| レイヤー | 技術 | バージョン | 選定理由 |
|---------|------|---------|---------|
| Desktop framework | Wails v2 | latest | Go+VueでネイティブデスクトップアプリになるElectron代替。HTTPサーバー不要、シングルバイナリ配布 |
| Backend | Go | 1.22+ | 子プロセス管理・goroutineによる並行ヘルスチェックに最適。cometの既知スタック |
| Router | chi | latest | 軽量・標準的。開発中のHTTP APIに使用（Wails devサーバーモード等） |
| ORM | GORM | latest | SQLiteサポート・シンプルなマイグレーション |
| Database | SQLite | via GORM | ローカルファースト・ゼロ設定・デスクトップアプリに最適 |
| Frontend | Vue 3 + Vite | latest | cometの主力スタック。Wails公式テンプレートあり |
| State | Pinia | latest | Vue 3標準。Wailsのイベント購読との相性良い |
| Styling | UnoCSS | latest | Tailwind互換プリセット・ビルド5〜10倍速い |
| Components | shadcn-vue | latest | テーブル・バッジ・ダイアログ等dev tool向けUIが揃う。コピペ方式で自由度高 |
| MCP Protocol | 自前実装 | — | JSON-RPC 2.0 over stdio + HTTP transport両対応 |
| Desktop binding | Wails bindings | — | GoメソッドをVueから直接呼ぶ（HTTPサーバー不要） |

---

## 不採用とその理由

| 技術 | 理由 |
|------|------|
| Nuxt 3 | SSR前提のためWailsと相性が悪い。デスクトップアプリにSSRは不要 |
| Next.js | 同上 |
| Electron | バンドルサイズ100MB超。Node.jsランタイムが肥大。Wailsで代替可能 |
| Tauri + Go sidecar | RustシェルにGoを同梱する構成が複雑。Wailsの方がシンプル |
| Tailwind CSS | UnoCSS（Tailwind互換）の方がビルドが高速。構文はほぼ同一 |
| Separate HTTP backend | デスクトップアプリにlocalhost HTTP serverは不要。Wails IPC/bindingsで代替 |

---

## MCP Protocol 方針

- **JSON-RPC 2.0** over **stdio** (subprocess) が主トランスポート
- **HTTP + SSE** トランスポートもv0.1.0で対応（Supabase MCP等がHTTP前提のため）
- 自前実装の理由: Go向けのMCPクライアントライブラリが未成熟。プロトコルがシンプルなので自前が最も制御しやすい

### 実装するMCPメソッド

| メソッド | 用途 |
|---------|------|
| `initialize` | ハンドシェイク・capability確認 |
| `tools/list` | ツール一覧取得 |
| `tools/call` | ツール実行 |

v0.2.0以降: `resources/list`, `resources/read`, `prompts/list`

---

## 配布方針

- **Wails build** → `.app` (macOS) / `.exe` (Windows) シングルバイナリ
- **GitHub Releases** でビルド済みアーティファクトを配布
- ビルドマトリクス: macOS arm64 / macOS amd64 / Windows amd64

---

## 将来の技術更新候補

| 項目 | 検討内容 |
|------|---------|
| Resources/Prompts API | v0.2.0でMCPの全メソッド対応 |
| Wails v3 | リリース後に移行検討（より軽量なアーキテクチャ） |
| SQLite → ファイルレスストレージ | 設定をJSONファイルで管理する方式も検討 |
