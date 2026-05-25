import { defineStore } from 'pinia'
import { ref } from 'vue'

export type Lang = 'ja' | 'en'

type Fn = (...args: unknown[]) => string
type Dict = Record<string, string | Fn>

const translations: Record<Lang, Dict> = {
  ja: {
    // Nav
    nav_servers: 'Servers',
    nav_import: 'Import',
    lang_toggle: 'EN',

    // ServersView
    sv_title: 'MCP Servers',
    sv_count: (n) => `${n} サーバー設定済み`,
    sv_import_btn: '↓ Claude Desktop をインポート',
    sv_importing: 'インポート中…',
    sv_add_btn: '+ サーバーを追加',
    sv_import_complete: 'インポート完了',
    sv_imported: (names) => `✓ 追加: ${names}`,
    sv_skipped: (names) => `— スキップ（登録済み）: ${names}`,
    sv_dismiss: '閉じる',
    sv_form_title: '新しいサーバー',
    sv_form_name: '名前',
    sv_form_transport: 'トランスポート',
    sv_form_command: 'コマンド',
    sv_form_args: 'Args（JSON 配列）',
    sv_form_url: 'URL',
    sv_form_add: '追加',
    sv_form_cancel: 'キャンセル',
    sv_empty_title: 'MCP サーバーがまだ登録されていません',
    sv_empty_desc: '以下の方法で MCP サーバーを追加してください',
    sv_step1_title: '既存の設定を取り込む（推奨）',
    sv_step1_desc: 'Claude Desktop や Claude Code を使っている場合は、設定を自動でインポートできます。',
    sv_step1_link: '→ インポートページへ',
    sv_step2_title: '手動でサーバーを追加',
    sv_step2_desc: 'コマンド（stdio）または URL（HTTP）でサーバーを直接登録します。',
    sv_step3_title: '接続してツールを確認',
    sv_step3_desc: 'サーバーを追加したら Connect を押して接続。Browse Tools でツール一覧とテスト実行ができます。',
    sv_import_config: '↓ 設定をインポート',
    sv_add_manual: '+ 手動で追加',
    sv_connecting: '接続中…',
    sv_connected: '● 接続済み',
    sv_retry: '⚠ 再試行',
    sv_connect: '接続',
    sv_delete: '削除',
    sv_delete_ok: '確認',
    sv_delete_confirm: (name) => `"${name}" を削除しますか？`,

    // ToolsView
    tv_error_title: '接続エラー',
    tv_error_none: 'エラー詳細なし',
    tv_degraded: '接続が不安定',
    tv_connecting: '接続中…',
    tv_connected_disconnect: '● 接続済み — 切断',
    tv_retry: '↺ 再接続',
    tv_connect: '接続',
    tv_delete_server: 'サーバーを削除',
    tv_tools: 'ツール',
    tv_loading: '読み込み中…',
    tv_connect_hint: '接続してツールを確認できます',
    tv_no_tools: 'ツールが見つかりません',
    tv_server_info: 'サーバー情報',
    tv_transport: 'Transport',
    tv_command: 'コマンド',
    tv_url: 'URL',
    tv_error_log: 'エラーログ',
    tv_arguments: 'Arguments',
    tv_edit_json: 'JSON で編集',
    tv_required: 'required',
    tv_arguments_json: 'Arguments (JSON)',
    tv_back_to_form: 'フォームに戻す',
    tv_executing: '実行中…',
    tv_execute: '▶  実行する',
    tv_error: 'エラー',
    tv_result: '結果',
    tv_delete_confirm: (name) => `"${name}" を削除しますか？`,

    // ImportView
    iv_title: 'MCP サーバーをインポート',
    iv_desc: '使用中のツールを選択して、登録済みの MCP サーバー設定をまとめて取り込みます。',
    iv_step1: 'Step 1 — 取り込み元を選択',
    iv_step2: 'Step 2 — 読み込み元ファイルの確認',
    iv_step2_desc_pre: '以下のパスのファイルから読み込みます。パスは',
    iv_step2_desc_bold: 'あなたの Mac のユーザー名',
    iv_step2_desc_post: 'を含むため人によって異なりますが、アプリが自動で設定します。違う場合は編集してください。',
    iv_detecting: '検出中…',
    iv_path_hint: '自動検出されたパスです。変更は不要です。',
    iv_browse: 'Browse…',
    iv_importing: 'インポート中…',
    iv_error_title: 'エラーが発生しました',
    iv_error_no_file: 'ファイルパスが指定されていません',
    iv_hint_no_file: 'ヒント: ファイルが存在しないパスです。Step 2 でパスを確認・修正してください。',
    iv_no_servers_title: 'MCP サーバーが見つかりませんでした',
    iv_no_servers_desc: 'ファイルは読めましたが mcpServers の設定がありません。',
    iv_read_file: '読み込んだファイル:',
    iv_try_custom: '「カスタム JSON ファイル」で別のパスを指定してお試しください。',
    iv_import_complete: 'インポート完了',
    iv_import_noop: 'インポート済み（変更なし）',
    iv_added: (n) => `追加 ${n}件`,
    iv_skipped_count: (n) => `スキップ（登録済み）${n}件`,
    iv_errors_count: (n) => `エラー ${n}件`,
    iv_go_servers: '→ Servers 画面で確認する',
    iv_import_btn: (method) => `${method} からインポート`,
    iv_custom_hint: 'mcpServers キーを含む JSON ファイルであれば取り込めます。',
    iv_claude_code_hint: (a, b) => `${a} が優先されます。なければ ${b} を使います。手動管理のファイルを使う場合はパスを変更してください。`,
    iv_cursor_hint: 'プロジェクトごとの .cursor/mcp.json を使う場合はパスを変更してください。',

    // Method titles
    iv_method_claude_desktop: 'Claude Desktop',
    iv_method_claude_desktop_desc: 'Claude Desktop アプリで設定したサーバーを取り込みます',
    iv_method_claude_code: 'Claude Code',
    iv_method_claude_code_desc: '~/.claude.json（または ~/.claude/claude_desktop_config.json）を取り込みます',
    iv_method_cursor: 'Cursor',
    iv_method_cursor_desc: '~/.cursor/mcp.json を取り込みます',
    iv_method_windsurf: 'Windsurf',
    iv_method_windsurf_desc: '~/.codeium/windsurf/mcp_config.json を取り込みます',
    iv_method_custom: 'カスタム JSON ファイル',
    iv_method_custom_desc: 'ファイルを直接指定して取り込みます',
  },

  en: {
    // Nav
    nav_servers: 'Servers',
    nav_import: 'Import',
    lang_toggle: 'JA',

    // ServersView
    sv_title: 'MCP Servers',
    sv_count: (n) => `${n} server(s) configured`,
    sv_import_btn: '↓ Import Claude Desktop',
    sv_importing: 'Importing…',
    sv_add_btn: '+ Add Server',
    sv_import_complete: 'Import complete',
    sv_imported: (names) => `✓ Imported: ${names}`,
    sv_skipped: (names) => `— Skipped (already exist): ${names}`,
    sv_dismiss: 'Dismiss',
    sv_form_title: 'New Server',
    sv_form_name: 'Name',
    sv_form_transport: 'Transport',
    sv_form_command: 'Command',
    sv_form_args: 'Args (JSON array)',
    sv_form_url: 'URL',
    sv_form_add: 'Add',
    sv_form_cancel: 'Cancel',
    sv_empty_title: 'No MCP servers configured yet',
    sv_empty_desc: 'Add an MCP server using one of the methods below',
    sv_step1_title: 'Import existing config (recommended)',
    sv_step1_desc: 'If you use Claude Desktop or Claude Code, import your servers automatically.',
    sv_step1_link: '→ Go to Import',
    sv_step2_title: 'Add server manually',
    sv_step2_desc: 'Register a server directly by command (stdio) or URL (HTTP).',
    sv_step3_title: 'Connect and explore tools',
    sv_step3_desc: 'After adding a server, click Connect. Then browse and test its tools.',
    sv_import_config: '↓ Import Config',
    sv_add_manual: '+ Add Manually',
    sv_connecting: 'Connecting…',
    sv_connected: '● Connected',
    sv_retry: '⚠ Retry',
    sv_connect: 'Connect',
    sv_delete: 'Delete',
    sv_delete_ok: 'Sure?',
    sv_delete_confirm: (name) => `Delete "${name}"?`,

    // ToolsView
    tv_error_title: 'Connection Error',
    tv_error_none: 'No error details',
    tv_degraded: 'Connection unstable',
    tv_connecting: 'Connecting…',
    tv_connected_disconnect: '● Connected — Disconnect',
    tv_retry: '↺ Retry Connect',
    tv_connect: 'Connect',
    tv_delete_server: 'Delete Server',
    tv_tools: 'Tools',
    tv_loading: 'Loading…',
    tv_connect_hint: 'Connect to view tools',
    tv_no_tools: 'No tools found',
    tv_server_info: 'Server Info',
    tv_transport: 'Transport',
    tv_command: 'Command',
    tv_url: 'URL',
    tv_error_log: 'Error Log',
    tv_arguments: 'Arguments',
    tv_edit_json: 'Edit as JSON',
    tv_required: 'required',
    tv_arguments_json: 'Arguments (JSON)',
    tv_back_to_form: 'Back to form',
    tv_executing: 'Running…',
    tv_execute: '▶  Execute',
    tv_error: 'Error',
    tv_result: 'Result',
    tv_delete_confirm: (name) => `Delete "${name}"?`,

    // ImportView
    iv_title: 'Import MCP Servers',
    iv_desc: 'Select your tool and import all registered MCP server configs at once.',
    iv_step1: 'Step 1 — Select source',
    iv_step2: 'Step 2 — Confirm file path',
    iv_step2_desc_pre: 'The path below includes your Mac ',
    iv_step2_desc_bold: 'username',
    iv_step2_desc_post: ' and will differ per user. The app sets it automatically — edit only if needed.',
    iv_detecting: 'Detecting…',
    iv_path_hint: 'Auto-detected. No changes needed.',
    iv_browse: 'Browse…',
    iv_importing: 'Importing…',
    iv_error_title: 'An error occurred',
    iv_error_no_file: 'No file path specified',
    iv_hint_no_file: 'Hint: The path does not exist. Check or correct the path in Step 2.',
    iv_no_servers_title: 'No MCP servers found',
    iv_no_servers_desc: 'File was read but contains no mcpServers configuration.',
    iv_read_file: 'File read:',
    iv_try_custom: 'Try specifying a different path using "Custom JSON File".',
    iv_import_complete: 'Import complete',
    iv_import_noop: 'Already up to date',
    iv_added: (n) => `Added ${n}`,
    iv_skipped_count: (n) => `Skipped (already exist) ${n}`,
    iv_errors_count: (n) => `Errors ${n}`,
    iv_go_servers: '→ View in Servers',
    iv_import_btn: (method) => `Import from ${method}`,
    iv_custom_hint: 'Any JSON file containing a mcpServers key is supported.',
    iv_claude_code_hint: (a, b) => `${a} takes priority. Falls back to ${b}. Change the path for a manually managed file.`,
    iv_cursor_hint: 'Change the path if you want to use a project-specific .cursor/mcp.json.',

    // Method titles
    iv_method_claude_desktop: 'Claude Desktop',
    iv_method_claude_desktop_desc: 'Import servers configured in the Claude Desktop app',
    iv_method_claude_code: 'Claude Code',
    iv_method_claude_code_desc: 'Import from ~/.claude.json (or ~/.claude/claude_desktop_config.json)',
    iv_method_cursor: 'Cursor',
    iv_method_cursor_desc: 'Import from ~/.cursor/mcp.json',
    iv_method_windsurf: 'Windsurf',
    iv_method_windsurf_desc: 'Import from ~/.codeium/windsurf/mcp_config.json',
    iv_method_custom: 'Custom JSON File',
    iv_method_custom_desc: 'Specify a file path directly',
  },
}

const safeStorage = typeof localStorage !== 'undefined' && typeof localStorage.getItem === 'function'
  ? localStorage
  : null

export const useI18nStore = defineStore('i18n', () => {
  const lang = ref<Lang>((safeStorage?.getItem('mcp_lang') as Lang) || 'ja')

  function setLang(l: Lang) {
    lang.value = l
    safeStorage?.setItem('mcp_lang', l)
  }

  function t(key: string, ...args: unknown[]): string {
    const dict = translations[lang.value]
    const val = dict[key]
    if (val === undefined) return key
    if (typeof val === 'function') return val(...args)
    return val
  }

  return { lang, setLang, t }
})
