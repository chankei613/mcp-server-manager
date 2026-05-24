<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useServersStore } from '@/stores/servers'
import { mcp } from '../../wailsjs/go/models'

const route = useRoute()
const router = useRouter()
const store = useServersStore()

const serverID = Number(route.params.id)
const server = computed(() => store.servers.find(s => s.ID === serverID))
const tools = ref<mcp.Tool[]>([])
const selectedTool = ref<mcp.Tool | null>(null)
const loadingTools = ref(false)
const connecting = ref(false)

const argsInput = ref('{}')
const executing = ref(false)
const result = ref<string | null>(null)
const execError = ref<string | null>(null)

const parsedArgs = computed<string[]>(() => {
  const raw = server.value?.args
  if (!raw) return []
  try { return JSON.parse(raw) } catch { return [raw] }
})

const fullCommand = computed(() => {
  if (!server.value) return ''
  const parts = [server.value.command, ...parsedArgs.value]
  return parts.join(' ')
})

const events = ref<{ event_type: string; message: string; CreatedAt: string }[]>([])

async function loadEvents() {
  try {
    const raw = await store.fetchEvents(serverID)
    events.value = raw.slice(0, 10) as typeof events.value
  } catch {
    // イベントログ取得失敗は無視
  }
}

const statusColors: Record<string, string> = {
  connected: 'bg-green-500',
  disconnected: 'bg-gray-400',
  error: 'bg-red-500',
  degraded: 'bg-yellow-500',
  connecting: 'bg-blue-500 animate-pulse',
}

async function loadTools() {
  loadingTools.value = true
  tools.value = []
  try {
    tools.value = await store.fetchTools(serverID)
  } catch {
    // エラーはサーバーステータスに反映される
  } finally {
    loadingTools.value = false
  }
}

async function handleConnect() {
  connecting.value = true
  try {
    await store.connect(serverID)
  } finally {
    connecting.value = false
  }
}

async function handleDisconnect() {
  await store.disconnect(serverID)
  tools.value = []
  selectedTool.value = null
}

async function handleDelete() {
  if (confirm(`"${server.value?.name}" を削除しますか？`)) {
    await store.removeServer(serverID)
    router.push('/servers')
  }
}

// immediate: true で既接続サーバーの初回マウント時にも発火
watch(() => server.value?.status, (newStatus, oldStatus) => {
  if (newStatus === 'connected') {
    loadTools()
    events.value = []
  } else if (oldStatus === 'connected' && newStatus !== 'connected') {
    tools.value = []
    selectedTool.value = null
  }
  if (newStatus === 'error') {
    loadEvents()
  }
}, { immediate: true })

function selectTool(tool: mcp.Tool) {
  selectedTool.value = tool
  result.value = null
  execError.value = null
  argsInput.value = generateDefaultArgs(tool.inputSchema as object)
}

function generateDefaultArgs(schema: object): string {
  const s = schema as { properties?: Record<string, { type: string }> }
  if (!s.properties) return '{}'
  const defaults: Record<string, unknown> = {}
  for (const [key, val] of Object.entries(s.properties)) {
    defaults[key] = val.type === 'string' ? '' : val.type === 'number' ? 0 : null
  }
  return JSON.stringify(defaults, null, 2)
}

async function execute() {
  if (!selectedTool.value) return
  executing.value = true
  result.value = null
  execError.value = null
  try {
    const res = await store.callTool(serverID, selectedTool.value.name, argsInput.value)
    result.value = JSON.stringify(res, null, 2)
  } catch (e) {
    execError.value = String(e)
  } finally {
    executing.value = false
  }
}
</script>

<template>
  <div class="flex h-full">
    <!-- Left panel: connect + tools list -->
    <div class="w-64 border-r border-border flex flex-col overflow-hidden shrink-0">

      <!-- Header -->
      <div class="p-3 border-b border-border flex items-center gap-2 shrink-0">
        <button
          @click="router.back()"
          class="text-muted-foreground hover:text-foreground text-sm p-1 rounded hover:bg-accent transition-colors"
        >←</button>
        <span class="w-2 h-2 rounded-full shrink-0" :class="statusColors[server?.status || 'disconnected']" />
        <h2 class="text-sm font-medium truncate flex-1">{{ server?.name }}</h2>
      </div>

      <!-- Connection control -->
      <div class="p-3 border-b border-border space-y-2 shrink-0">
        <!-- Error banner -->
        <div v-if="server?.status === 'error'" class="p-2.5 rounded-lg bg-red-50 border border-red-200">
          <p class="text-xs font-semibold text-red-800 mb-1">接続エラー</p>
          <p class="text-xs text-red-700 font-mono break-all leading-relaxed">
            {{ server.last_error || 'エラー詳細なし' }}
          </p>
        </div>
        <div v-else-if="server?.status === 'degraded'" class="p-2.5 rounded-lg bg-yellow-50 border border-yellow-200">
          <p class="text-xs font-semibold text-yellow-800">接続が不安定</p>
          <p v-if="server.last_error" class="text-xs text-yellow-700 mt-0.5 break-all">{{ server.last_error }}</p>
        </div>

        <!-- Toggle button -->
        <button
          v-if="server?.status === 'connecting' || connecting"
          disabled
          class="w-full py-2 text-xs rounded-lg bg-blue-50 text-blue-600 border border-blue-200 cursor-not-allowed"
        >Connecting…</button>
        <button
          v-else-if="server?.status === 'connected'"
          @click="handleDisconnect"
          class="w-full py-2 text-xs rounded-lg bg-green-50 text-green-700 border border-green-300 hover:bg-green-100 transition-colors"
        >● Connected — Disconnect</button>
        <button
          v-else-if="server?.status === 'error'"
          @click="handleConnect"
          :disabled="connecting"
          class="w-full py-2 text-xs rounded-lg bg-red-600 text-white hover:bg-red-700 transition-colors font-medium disabled:opacity-50"
        >↺ Retry Connect</button>
        <button
          v-else
          @click="handleConnect"
          :disabled="connecting"
          class="w-full py-2 text-xs rounded-lg bg-gray-900 text-white hover:bg-gray-700 transition-colors font-medium disabled:opacity-50"
        >Connect</button>

        <!-- Delete -->
        <button
          @click="handleDelete"
          class="w-full py-1.5 text-xs text-destructive border border-destructive/30 rounded-lg hover:bg-destructive/10 transition-colors"
        >Delete Server</button>
      </div>

      <!-- Tools list -->
      <div class="flex-1 overflow-auto flex flex-col">
        <div class="px-3 py-2 border-b border-border shrink-0">
          <p class="text-xs font-semibold text-muted-foreground uppercase tracking-wide">
            Tools<span v-if="tools.length" class="ml-1 normal-case font-normal">({{ tools.length }})</span>
          </p>
        </div>
        <div v-if="loadingTools" class="p-4 text-xs text-muted-foreground text-center">Loading…</div>
        <div v-else-if="server?.status !== 'connected'" class="p-4 text-xs text-muted-foreground text-center leading-relaxed">
          Connect してツールを確認できます
        </div>
        <div v-else-if="tools.length === 0" class="p-4 text-xs text-muted-foreground text-center">ツールが見つかりません</div>
        <div v-else class="p-2">
          <button
            v-for="tool in tools"
            :key="tool.name"
            @click="selectTool(tool)"
            class="w-full text-left px-3 py-2 rounded-md hover:bg-accent transition-colors mb-0.5"
            :class="selectedTool?.name === tool.name ? 'bg-accent' : ''"
          >
            <p class="text-xs font-medium text-foreground truncate">{{ tool.name }}</p>
            <p class="text-xs text-muted-foreground truncate mt-0.5">{{ tool.description }}</p>
          </button>
        </div>
      </div>
    </div>

    <!-- Right panel -->
    <div class="flex-1 flex flex-col overflow-hidden">

      <!-- Empty state: server info -->
      <template v-if="!selectedTool">
        <div class="flex-1 overflow-auto p-6">
          <div class="max-w-md">
            <h3 class="text-sm font-semibold mb-4 text-foreground">Server Info</h3>
            <div class="space-y-4 text-sm">
              <div class="flex gap-3">
                <span class="text-muted-foreground w-24 shrink-0">Transport</span>
                <span class="font-mono">{{ server?.transport }}</span>
              </div>
              <template v-if="server?.transport === 'stdio'">
                <div>
                  <span class="text-muted-foreground block mb-1">Command</span>
                  <div class="font-mono bg-gray-50 border border-gray-200 rounded px-3 py-2 text-xs text-gray-800 break-all leading-relaxed select-all">{{ fullCommand }}</div>
                </div>
              </template>
              <div v-if="server?.transport === 'http'">
                <span class="text-muted-foreground block mb-1">URL</span>
                <div class="font-mono bg-gray-50 border border-gray-200 rounded px-3 py-2 text-xs text-gray-800 break-all select-all">{{ server?.url }}</div>
              </div>

              <!-- エラーログ -->
              <template v-if="server?.last_error || events.length">
                <div>
                  <span class="text-muted-foreground block mb-2">Error Log</span>
                  <div class="bg-red-50 border border-red-200 rounded overflow-hidden">
                    <div class="px-3 py-2 border-b border-red-200 bg-red-100">
                      <p class="text-xs font-mono text-red-800 break-all">{{ server?.last_error }}</p>
                    </div>
                    <div v-if="events.length" class="divide-y divide-red-100 max-h-48 overflow-auto">
                      <div
                        v-for="(ev, i) in events"
                        :key="i"
                        class="px-3 py-1.5 text-xs font-mono"
                        :class="ev.event_type === 'error' ? 'text-red-700' : 'text-gray-600'"
                      >
                        <span class="opacity-60 mr-2">{{ ev.event_type }}</span>{{ ev.message }}
                      </div>
                    </div>
                  </div>
                </div>
              </template>
            </div>
          </div>
        </div>
      </template>

      <!-- Tool execution -->
      <template v-else>
        <div class="p-4 border-b border-border shrink-0">
          <h3 class="text-sm font-semibold">{{ selectedTool.name }}</h3>
          <p class="text-xs text-muted-foreground mt-0.5">{{ selectedTool.description }}</p>
        </div>
        <div class="flex-1 flex flex-col p-4 gap-4 overflow-auto">
          <div>
            <label class="text-xs text-muted-foreground mb-1 block">Arguments (JSON)</label>
            <textarea
              v-model="argsInput"
              rows="8"
              class="w-full p-3 text-xs font-mono bg-input border border-border rounded-md focus:outline-none focus:ring-1 focus:ring-ring resize-none"
            />
          </div>
          <button
            @click="execute"
            :disabled="executing"
            class="self-start px-4 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:opacity-90 disabled:opacity-50 transition-opacity"
          >{{ executing ? 'Executing…' : '▶ Execute' }}</button>
          <div v-if="result || execError">
            <label class="text-xs text-muted-foreground mb-1 block">{{ execError ? 'Error' : 'Result' }}</label>
            <pre
              class="p-3 text-xs font-mono rounded-md border overflow-auto max-h-64"
              :class="execError ? 'border-destructive/50 text-destructive bg-destructive/10' : 'border-border bg-muted/20'"
            >{{ execError || result }}</pre>
          </div>
        </div>
      </template>

    </div>
  </div>
</template>
