<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useServersStore } from '@/stores/servers'

const store = useServersStore()
const router = useRouter()

type Method = 'claude-desktop' | 'claude-code' | 'cursor' | 'windsurf' | 'custom'

const selectedMethod = ref<Method>('claude-desktop')
const claudeDesktopPath = ref('')
const claudeCodePath = ref('')
const cursorPath = ref('')
const windsurfPath = ref('')
const customPath = ref('')
const importing = ref(false)
const result = ref<{ imported: string[], skipped: string[], errors: string[] } | null>(null)
const error = ref<string | null>(null)
const lastImportedPath = ref('')

onMounted(async () => {
  try {
    const { GetClaudeDesktopConfigPath } = await import('../../wailsjs/go/main/App')
    claudeDesktopPath.value = await GetClaudeDesktopConfigPath()
  } catch { /* wails dev 起動前は空になる */ }
  claudeCodePath.value = await store.getLocalClaudePath()
  cursorPath.value = await store.getCursorConfigPath()
  windsurfPath.value = await store.getWindsurfConfigPath()
})

async function handlePickFile() {
  const picked = await store.pickFile()
  if (picked) customPath.value = picked
}

function currentPath(): string {
  if (selectedMethod.value === 'claude-desktop') return claudeDesktopPath.value
  if (selectedMethod.value === 'claude-code') return claudeCodePath.value
  if (selectedMethod.value === 'cursor') return cursorPath.value
  if (selectedMethod.value === 'windsurf') return windsurfPath.value
  return customPath.value
}

async function handleImport() {
  const path = currentPath()
  if (!path) {
    error.value = 'ファイルパスが指定されていません'
    return
  }
  importing.value = true
  result.value = null
  error.value = null
  lastImportedPath.value = path
  try {
    result.value = await store.importFromPath(path)
  } catch (e) {
    error.value = String(e)
  } finally {
    importing.value = false
  }
}

function goToServers() {
  router.push('/servers')
}

const methodLabels: Record<Method, { title: string; desc: string }> = {
  'claude-desktop': { title: 'Claude Desktop', desc: 'Claude Desktop アプリで設定したサーバーを取り込みます' },
  'claude-code':    { title: 'Claude Code', desc: '~/.claude.json（または ~/.claude/claude_desktop_config.json）を取り込みます' },
  'cursor':         { title: 'Cursor', desc: '~/.cursor/mcp.json を取り込みます' },
  'windsurf':       { title: 'Windsurf', desc: '~/.codeium/windsurf/mcp_config.json を取り込みます' },
  'custom':         { title: 'カスタム JSON ファイル', desc: 'ファイルを直接指定して取り込みます' },
}
</script>

<template>
  <div class="max-w-2xl mx-auto p-6">
    <!-- Header -->
    <div class="mb-6">
      <h2 class="text-xl font-semibold text-foreground">MCP サーバーをインポート</h2>
      <p class="text-sm text-muted-foreground mt-1">
        使用中のツールを選択して、登録済みの MCP サーバー設定をまとめて取り込みます。
      </p>
    </div>

    <!-- Step 1: Method selection -->
    <div class="mb-5">
      <p class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-3">
        Step 1 — 取り込み元を選択
      </p>

      <div class="space-y-2">
        <!-- Claude Desktop -->
        <label
          class="flex items-center gap-3 p-3.5 border-2 rounded-lg cursor-pointer transition-all"
          :class="selectedMethod === 'claude-desktop'
            ? 'border-blue-500 bg-blue-50'
            : 'border-border bg-white hover:border-blue-200 hover:bg-blue-50/30'"
        >
          <input type="radio" v-model="selectedMethod" value="claude-desktop" class="sr-only" />
          <span
            class="flex items-center justify-center w-5 h-5 rounded-full border-2 shrink-0 transition-all"
            :class="selectedMethod === 'claude-desktop' ? 'border-blue-500 bg-blue-500' : 'border-muted-foreground'"
          >
            <svg v-if="selectedMethod === 'claude-desktop'" class="w-3 h-3 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3.5">
              <polyline points="20 6 9 17 4 12" />
            </svg>
          </span>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium" :class="selectedMethod === 'claude-desktop' ? 'text-blue-900' : 'text-foreground'">
              Claude Desktop
            </p>
            <p class="text-xs text-muted-foreground">Claude Desktop アプリで設定したサーバーを取り込みます</p>
          </div>
        </label>

        <!-- Claude Code -->
        <label
          class="flex items-center gap-3 p-3.5 border-2 rounded-lg cursor-pointer transition-all"
          :class="selectedMethod === 'claude-code'
            ? 'border-blue-500 bg-blue-50'
            : 'border-border bg-white hover:border-blue-200 hover:bg-blue-50/30'"
        >
          <input type="radio" v-model="selectedMethod" value="claude-code" class="sr-only" />
          <span
            class="flex items-center justify-center w-5 h-5 rounded-full border-2 shrink-0 transition-all"
            :class="selectedMethod === 'claude-code' ? 'border-blue-500 bg-blue-500' : 'border-muted-foreground'"
          >
            <svg v-if="selectedMethod === 'claude-code'" class="w-3 h-3 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3.5">
              <polyline points="20 6 9 17 4 12" />
            </svg>
          </span>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium" :class="selectedMethod === 'claude-code' ? 'text-blue-900' : 'text-foreground'">
              Claude Code
            </p>
            <p class="text-xs text-muted-foreground">~/.claude.json（または ~/.claude/claude_desktop_config.json）を取り込みます</p>
          </div>
        </label>

        <!-- Cursor -->
        <label
          class="flex items-center gap-3 p-3.5 border-2 rounded-lg cursor-pointer transition-all"
          :class="selectedMethod === 'cursor'
            ? 'border-blue-500 bg-blue-50'
            : 'border-border bg-white hover:border-blue-200 hover:bg-blue-50/30'"
        >
          <input type="radio" v-model="selectedMethod" value="cursor" class="sr-only" />
          <span
            class="flex items-center justify-center w-5 h-5 rounded-full border-2 shrink-0 transition-all"
            :class="selectedMethod === 'cursor' ? 'border-blue-500 bg-blue-500' : 'border-muted-foreground'"
          >
            <svg v-if="selectedMethod === 'cursor'" class="w-3 h-3 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3.5">
              <polyline points="20 6 9 17 4 12" />
            </svg>
          </span>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium" :class="selectedMethod === 'cursor' ? 'text-blue-900' : 'text-foreground'">Cursor</p>
            <p class="text-xs text-muted-foreground">~/.cursor/mcp.json を取り込みます</p>
          </div>
        </label>

        <!-- Windsurf -->
        <label
          class="flex items-center gap-3 p-3.5 border-2 rounded-lg cursor-pointer transition-all"
          :class="selectedMethod === 'windsurf'
            ? 'border-blue-500 bg-blue-50'
            : 'border-border bg-white hover:border-blue-200 hover:bg-blue-50/30'"
        >
          <input type="radio" v-model="selectedMethod" value="windsurf" class="sr-only" />
          <span
            class="flex items-center justify-center w-5 h-5 rounded-full border-2 shrink-0 transition-all"
            :class="selectedMethod === 'windsurf' ? 'border-blue-500 bg-blue-500' : 'border-muted-foreground'"
          >
            <svg v-if="selectedMethod === 'windsurf'" class="w-3 h-3 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3.5">
              <polyline points="20 6 9 17 4 12" />
            </svg>
          </span>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium" :class="selectedMethod === 'windsurf' ? 'text-blue-900' : 'text-foreground'">Windsurf</p>
            <p class="text-xs text-muted-foreground">~/.codeium/windsurf/mcp_config.json を取り込みます</p>
          </div>
        </label>

        <!-- Custom -->
        <label
          class="flex items-center gap-3 p-3.5 border-2 rounded-lg cursor-pointer transition-all"
          :class="selectedMethod === 'custom'
            ? 'border-blue-500 bg-blue-50'
            : 'border-border bg-white hover:border-blue-200 hover:bg-blue-50/30'"
        >
          <input type="radio" v-model="selectedMethod" value="custom" class="sr-only" />
          <span
            class="flex items-center justify-center w-5 h-5 rounded-full border-2 shrink-0 transition-all"
            :class="selectedMethod === 'custom' ? 'border-blue-500 bg-blue-500' : 'border-muted-foreground'"
          >
            <svg v-if="selectedMethod === 'custom'" class="w-3 h-3 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3.5">
              <polyline points="20 6 9 17 4 12" />
            </svg>
          </span>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium" :class="selectedMethod === 'custom' ? 'text-blue-900' : 'text-foreground'">カスタム JSON ファイル</p>
            <p class="text-xs text-muted-foreground">ファイルを直接指定して取り込みます</p>
          </div>
        </label>
      </div>
    </div>

    <!-- Step 2: Path confirmation -->
    <div class="mb-6 p-4 border border-border rounded-lg bg-white space-y-3">
      <div>
        <p class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-0.5">
          Step 2 — 読み込み元ファイルの確認
        </p>
        <p class="text-xs text-muted-foreground">
          以下のパスのファイルから読み込みます。パスは<strong>あなたの Mac のユーザー名</strong>を含むため人によって異なりますが、アプリが自動で設定します。違う場合は編集してください。
        </p>
      </div>

      <!-- Claude Desktop -->
      <template v-if="selectedMethod === 'claude-desktop'">
        <code class="block w-full px-3 py-2 text-xs bg-gray-100 border border-border rounded-md text-gray-700 break-all leading-relaxed select-all">
          {{ claudeDesktopPath || '検出中…' }}
        </code>
        <p class="text-xs text-gray-500">自動検出されたパスです。変更は不要です。</p>
      </template>

      <!-- Claude Code -->
      <template v-else-if="selectedMethod === 'claude-code'">
        <input
          v-model="claudeCodePath"
          placeholder="~/.claude.json"
          class="w-full px-3 py-2 text-sm bg-input border border-border rounded-md font-mono focus:outline-none focus:ring-1 focus:ring-ring"
        />
        <code class="block text-xs text-gray-500 break-all leading-relaxed px-1">{{ claudeCodePath }}</code>
        <p class="text-xs text-gray-500">
          <code class="bg-gray-100 px-1 rounded">~/.claude.json</code> が優先されます。なければ <code class="bg-gray-100 px-1 rounded">~/.claude/claude_desktop_config.json</code> を使います。手動管理のファイルを使う場合はパスを変更してください。
        </p>
      </template>

      <!-- Cursor -->
      <template v-else-if="selectedMethod === 'cursor'">
        <input
          v-model="cursorPath"
          placeholder="~/.cursor/mcp.json"
          class="w-full px-3 py-2 text-sm bg-input border border-border rounded-md font-mono focus:outline-none focus:ring-1 focus:ring-ring"
        />
        <code class="block text-xs text-gray-500 break-all leading-relaxed px-1">{{ cursorPath }}</code>
        <p class="text-xs text-gray-500">プロジェクトごとの <code class="bg-gray-100 px-1 rounded">.cursor/mcp.json</code> を使う場合はパスを変更してください。</p>
      </template>

      <!-- Windsurf -->
      <template v-else-if="selectedMethod === 'windsurf'">
        <input
          v-model="windsurfPath"
          placeholder="~/.codeium/windsurf/mcp_config.json"
          class="w-full px-3 py-2 text-sm bg-input border border-border rounded-md font-mono focus:outline-none focus:ring-1 focus:ring-ring"
        />
        <code class="block text-xs text-gray-500 break-all leading-relaxed px-1">{{ windsurfPath }}</code>
      </template>

      <!-- Custom -->
      <template v-else>
        <div class="flex gap-2">
          <input
            v-model="customPath"
            placeholder="/path/to/claude_desktop_config.json"
            class="flex-1 px-3 py-2 text-sm bg-input border border-border rounded-md font-mono focus:outline-none focus:ring-1 focus:ring-ring"
          />
          <button
            @click="handlePickFile"
            class="px-3 py-2 text-sm border border-border rounded-md hover:bg-accent transition-colors shrink-0"
          >
            Browse…
          </button>
        </div>
        <code v-if="customPath" class="block text-xs text-gray-500 break-all leading-relaxed px-1">{{ customPath }}</code>
        <p class="text-xs text-gray-500">
          <code class="bg-gray-100 px-1 rounded">mcpServers</code> キーを含む JSON ファイルであれば取り込めます。
        </p>
      </template>
    </div>

    <!-- Loading feedback (ボタン上) -->
    <div v-if="importing" class="mb-4 p-4 border border-blue-200 bg-blue-50 rounded-xl flex items-start gap-3">
      <svg class="w-5 h-5 text-blue-600 animate-spin shrink-0 mt-0.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M21 12a9 9 0 1 1-6.219-8.56" />
      </svg>
      <div class="min-w-0">
        <p class="text-sm font-medium text-blue-900">インポート中…</p>
        <p class="text-xs text-blue-700 font-mono mt-1 break-all">{{ currentPath() }}</p>
      </div>
    </div>

    <!-- Error (ボタン上) -->
    <div v-if="error" class="mb-4 p-4 border border-red-200 bg-red-50 rounded-xl text-sm text-red-800">
      <p class="font-semibold">エラーが発生しました</p>
      <p class="mt-1 font-mono text-xs break-all">{{ error }}</p>
      <p v-if="error.includes('no such file')" class="mt-2 text-xs text-red-700">
        ヒント: ファイルが存在しないパスです。Step 2 でパスを確認・修正してください。
      </p>
    </div>

    <!-- Result (ボタン上) -->
    <div v-if="result && !importing" class="mb-4 rounded-xl border overflow-hidden">
      <!-- 0件: mcpServers なし -->
      <template v-if="!result.imported?.length && !result.skipped?.length && !result.errors?.length">
        <div class="p-4 bg-yellow-50 border-b border-yellow-200">
          <p class="text-sm font-semibold text-yellow-900">MCP サーバーが見つかりませんでした</p>
        </div>
        <div class="p-4 bg-white space-y-2 text-sm text-gray-700">
          <p>ファイルは読めましたが <code class="bg-gray-100 px-1 rounded text-xs">mcpServers</code> の設定がありません。</p>
          <p class="text-xs text-gray-500">読み込んだファイル:</p>
          <p class="text-xs font-mono text-gray-700 break-all bg-gray-50 px-2 py-1 rounded">{{ lastImportedPath }}</p>
          <p class="text-xs text-gray-500">「カスタム JSON ファイル」で別のパスを指定してお試しください。</p>
        </div>
      </template>

      <!-- 成功 / スキップ -->
      <template v-else>
        <div class="p-4 border-b" :class="result.imported?.length ? 'bg-green-50 border-green-200' : 'bg-gray-50 border-gray-200'">
          <p class="text-sm font-semibold" :class="result.imported?.length ? 'text-green-900' : 'text-gray-700'">
            {{ result.imported?.length ? 'インポート完了' : 'インポート済み（変更なし）' }}
          </p>
          <p class="text-xs font-mono mt-1 break-all" :class="result.imported?.length ? 'text-green-700' : 'text-gray-500'">
            {{ lastImportedPath }}
          </p>
        </div>
        <div class="p-4 bg-white space-y-3">
          <div v-if="result.imported?.length" class="space-y-1">
            <p class="text-xs font-semibold text-green-700 uppercase tracking-wide">追加 {{ result.imported.length }}件</p>
            <ul class="space-y-1">
              <li v-for="name in result.imported" :key="name" class="flex items-center gap-2">
                <span class="text-green-600 font-bold text-sm">✓</span>
                <span class="font-mono text-xs">{{ name }}</span>
              </li>
            </ul>
          </div>

          <div v-if="result.skipped?.length" class="space-y-1">
            <p class="text-xs font-semibold text-gray-500 uppercase tracking-wide">スキップ（登録済み）{{ result.skipped.length }}件</p>
            <ul class="space-y-1">
              <li v-for="name in result.skipped" :key="name" class="flex items-center gap-2 text-gray-500">
                <span class="text-sm">—</span>
                <span class="font-mono text-xs">{{ name }}</span>
              </li>
            </ul>
          </div>

          <div v-if="result.errors?.length" class="space-y-1">
            <p class="text-xs font-semibold text-red-600 uppercase tracking-wide">エラー {{ result.errors.length }}件</p>
            <ul class="space-y-1">
              <li v-for="err in result.errors" :key="err" class="flex items-start gap-2 text-red-700">
                <span class="shrink-0 text-sm">✗</span>
                <span class="text-xs">{{ err }}</span>
              </li>
            </ul>
          </div>

          <button
            v-if="result.imported?.length"
            @click="goToServers"
            class="w-full py-2 text-sm font-medium bg-gray-900 text-white rounded-lg hover:bg-gray-700 transition-colors"
          >
            → Servers 画面で確認する
          </button>
        </div>
      </template>
    </div>

    <!-- Step 3: Import button -->
    <button
      @click="handleImport"
      :disabled="importing || (selectedMethod === 'custom' && !customPath)"
      class="w-full flex items-center justify-center gap-2 py-3 text-base font-semibold bg-blue-600 text-white rounded-xl shadow-sm hover:bg-blue-700 active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed transition-all"
    >
      <svg v-if="!importing" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
        <polyline points="7 10 12 15 17 10" />
        <line x1="12" y1="15" x2="12" y2="3" />
      </svg>
      <svg v-else class="w-5 h-5 animate-spin" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M21 12a9 9 0 1 1-6.219-8.56" />
      </svg>
      {{ importing ? 'インポート中…' : `${methodLabels[selectedMethod].title} からインポート` }}
    </button>
  </div>
</template>
