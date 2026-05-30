<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useServersStore } from '@/stores/servers'
import { useI18nStore } from '@/stores/i18n'
import { mcp } from '../../wailsjs/go/models'

const route = useRoute()
const router = useRouter()
const store = useServersStore()
const i18n = useI18nStore()

const serverID = Number(route.params.id)
const server = computed(() => store.servers.find(s => s.ID === serverID))
const tools = ref<mcp.Tool[]>([])
const selectedTool = ref<mcp.Tool | null>(null)
const loadingTools = ref(false)
const connecting = ref(false)

const argsInput = ref('{}')
const formValues = ref<Record<string, unknown>>({})
const showRawJson = ref(false)
const executing = ref(false)
const result = ref<string | null>(null)
const execError = ref<string | null>(null)

interface SchemaProp {
  key: string
  type: string
  description: string
  required: boolean
  enum?: string[]
}

const schemaProperties = computed<SchemaProp[]>(() => {
  if (!selectedTool.value) return []
  const schema = selectedTool.value.inputSchema as {
    properties?: Record<string, { type?: string; description?: string; enum?: string[] }>
    required?: string[]
  }
  if (!schema?.properties) return []
  const required = schema.required ?? []
  return Object.entries(schema.properties).map(([key, prop]) => ({
    key,
    type: prop.type ?? 'string',
    description: prop.description ?? '',
    required: required.includes(key),
    enum: prop.enum,
  }))
})

const hasSchema = computed(() => schemaProperties.value.length > 0)

function syncFormToJson() {
  const result: Record<string, unknown> = {}
  for (const prop of schemaProperties.value) {
    const val = formValues.value[prop.key]
    if (prop.type === 'object' || prop.type === 'array') {
      try { result[prop.key] = JSON.parse(val as string) } catch { result[prop.key] = val }
    } else {
      result[prop.key] = val
    }
  }
  argsInput.value = JSON.stringify(result, null, 2)
}

const parsedArgs = computed<string[]>(() => {
  const raw = server.value?.args
  if (!raw) return []
  try { return JSON.parse(raw) } catch { return [raw] }
})

const fullCommand = computed(() => {
  if (!server.value) return ''
  return [server.value.command, ...parsedArgs.value].join(' ')
})

const events = ref<{ event_type: string; message: string; CreatedAt: string }[]>([])

async function loadEvents() {
  try {
    const raw = await store.fetchEvents(serverID)
    events.value = raw.slice(0, 10) as typeof events.value
  } catch { /* ignore */ }
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
  try { tools.value = await store.fetchTools(serverID) }
  catch { /* status reflects error */ }
  finally { loadingTools.value = false }
}

async function handleConnect() {
  connecting.value = true
  try { await store.connect(serverID) }
  finally { connecting.value = false }
}

async function handleDisconnect() {
  await store.disconnect(serverID)
  tools.value = []
  selectedTool.value = null
}

const pendingDelete = ref(false)
const deleted = ref(false)

async function handleDelete() {
  if (!pendingDelete.value) {
    pendingDelete.value = true
    return
  }
  pendingDelete.value = false
  await store.removeServer(serverID)
  deleted.value = true
  setTimeout(() => router.push('/servers'), 1500)
}

watch(() => server.value?.status, (newStatus, oldStatus) => {
  if (newStatus === 'connected') { loadTools(); events.value = [] }
  else if (oldStatus === 'connected' && newStatus !== 'connected') { tools.value = []; selectedTool.value = null }
  if (newStatus === 'error') loadEvents()
}, { immediate: true })

function selectTool(tool: mcp.Tool) {
  selectedTool.value = tool
  result.value = null
  execError.value = null
  showRawJson.value = false

  const schema = tool.inputSchema as {
    properties?: Record<string, { type?: string; default?: unknown }>
    required?: string[]
  }
  const values: Record<string, unknown> = {}
  for (const [key, prop] of Object.entries(schema?.properties ?? {})) {
    if (prop.default !== undefined) {
      values[key] = prop.default
    } else if (prop.type === 'number' || prop.type === 'integer') {
      values[key] = 0
    } else if (prop.type === 'boolean') {
      values[key] = false
    } else if (prop.type === 'object') {
      values[key] = '{}'
    } else if (prop.type === 'array') {
      values[key] = '[]'
    } else {
      values[key] = ''
    }
  }
  formValues.value = values
  syncFormToJson()
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
  <!-- Deleted overlay -->
  <Transition name="fade">
    <div v-if="deleted" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
      <div class="bg-white rounded-2xl shadow-xl px-10 py-8 flex flex-col items-center gap-3">
        <div class="w-12 h-12 rounded-full bg-green-100 flex items-center justify-center">
          <svg class="w-6 h-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <p class="text-sm font-semibold text-foreground">{{ i18n.lang === 'ja' ? '削除しました' : 'Deleted' }}</p>
      </div>
    </div>
  </Transition>

  <div class="flex h-full">
    <!-- Left panel -->
    <div class="w-64 border-r border-border flex flex-col overflow-hidden shrink-0">

      <!-- Header -->
      <div class="p-3 border-b border-border flex items-center gap-2 shrink-0">
        <button @click="router.back()" class="text-muted-foreground hover:text-foreground text-sm p-1 rounded hover:bg-accent transition-colors">←</button>
        <span class="w-2 h-2 rounded-full shrink-0" :class="statusColors[server?.status || 'disconnected']" />
        <h2 class="text-sm font-medium truncate flex-1">{{ server?.name }}</h2>
      </div>

      <!-- Connection control -->
      <div class="p-3 border-b border-border space-y-2 shrink-0">
        <div v-if="server?.status === 'error'" class="p-2.5 rounded-lg bg-red-50 border border-red-200">
          <p class="text-xs font-semibold text-red-800 mb-1">{{ i18n.t('tv_error_title') }}</p>
          <p class="text-xs text-red-700 font-mono break-all leading-relaxed">{{ server.last_error || i18n.t('tv_error_none') }}</p>
        </div>
        <div v-else-if="server?.status === 'degraded'" class="p-2.5 rounded-lg bg-yellow-50 border border-yellow-200">
          <p class="text-xs font-semibold text-yellow-800">{{ i18n.t('tv_degraded') }}</p>
          <p v-if="server.last_error" class="text-xs text-yellow-700 mt-0.5 break-all">{{ server.last_error }}</p>
        </div>

        <button v-if="server?.status === 'connecting' || connecting" disabled
          class="w-full py-2 text-xs rounded-lg bg-blue-50 text-blue-600 border border-blue-200 cursor-not-allowed">{{ i18n.t('tv_connecting') }}</button>
        <button v-else-if="server?.status === 'connected'" @click="handleDisconnect"
          class="w-full py-2 text-xs rounded-lg bg-green-50 text-green-700 border border-green-200 hover:bg-green-100 transition-colors">{{ i18n.t('tv_connected_disconnect') }}</button>
        <button v-else-if="server?.status === 'error'" @click="handleConnect" :disabled="connecting"
          class="w-full py-2 text-xs rounded-lg bg-red-600 text-white hover:bg-red-700 transition-colors font-medium disabled:opacity-50">{{ i18n.t('tv_retry') }}</button>
        <button v-else @click="handleConnect" :disabled="connecting"
          class="w-full py-2 text-xs rounded-lg bg-gray-900 text-white hover:bg-gray-700 transition-colors font-medium disabled:opacity-50">{{ i18n.t('tv_connect') }}</button>

        <button @click="handleDelete"
          :class="pendingDelete
            ? 'w-full py-1.5 text-xs bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors font-medium'
            : 'w-full py-1.5 text-xs text-red-500 border border-red-300 rounded-lg hover:bg-red-50 transition-colors'"
        >{{ pendingDelete ? (i18n.lang === 'ja' ? '本当に削除する' : 'Confirm Delete') : i18n.t('tv_delete_server') }}</button>
      </div>

      <!-- Tools list -->
      <div class="flex-1 overflow-auto flex flex-col">
        <div class="px-3 py-2 border-b border-border shrink-0">
          <p class="text-xs font-semibold text-muted-foreground uppercase tracking-wide">
            {{ i18n.t('tv_tools') }}<span v-if="tools.length" class="ml-1 normal-case font-normal">({{ tools.length }})</span>
          </p>
        </div>
        <div v-if="loadingTools" class="p-4 text-xs text-muted-foreground text-center">{{ i18n.t('tv_loading') }}</div>
        <div v-else-if="server?.status !== 'connected'" class="p-4 text-xs text-muted-foreground text-center leading-relaxed">{{ i18n.t('tv_connect_hint') }}</div>
        <div v-else-if="tools.length === 0" class="p-4 text-xs text-muted-foreground text-center">{{ i18n.t('tv_no_tools') }}</div>
        <div v-else class="p-2">
          <button v-for="tool in tools" :key="tool.name" @click="selectTool(tool)"
            class="w-full text-left px-3 py-2 rounded-md transition-colors mb-0.5"
            :class="selectedTool?.name === tool.name
              ? 'bg-gray-900 text-white'
              : 'hover:bg-gray-100 text-gray-700'">
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
            <h3 class="text-sm font-semibold mb-4 text-foreground">{{ i18n.t('tv_server_info') }}</h3>
            <div class="space-y-4 text-sm">
              <div class="flex gap-3">
                <span class="text-muted-foreground w-24 shrink-0">{{ i18n.t('tv_transport') }}</span>
                <span class="font-mono">{{ server?.transport }}</span>
              </div>
              <template v-if="server?.transport === 'stdio'">
                <div>
                  <span class="text-muted-foreground block mb-1">{{ i18n.t('tv_command') }}</span>
                  <div class="font-mono bg-gray-50 border border-gray-200 rounded px-3 py-2 text-xs text-gray-800 break-all leading-relaxed select-all">{{ fullCommand }}</div>
                </div>
              </template>
              <div v-if="server?.transport === 'http'">
                <span class="text-muted-foreground block mb-1">{{ i18n.t('tv_url') }}</span>
                <div class="font-mono bg-gray-50 border border-gray-200 rounded px-3 py-2 text-xs text-gray-800 break-all select-all">{{ server?.url }}</div>
              </div>
              <template v-if="server?.last_error || events.length">
                <div>
                  <span class="text-muted-foreground block mb-2">{{ i18n.t('tv_error_log') }}</span>
                  <div class="bg-red-50 border border-red-200 rounded overflow-hidden">
                    <div class="px-3 py-2 border-b border-red-200 bg-red-100">
                      <p class="text-xs font-mono text-red-800 break-all">{{ server?.last_error }}</p>
                    </div>
                    <div v-if="events.length" class="divide-y divide-red-100 max-h-48 overflow-auto">
                      <div v-for="(ev, i) in events" :key="i" class="px-3 py-1.5 text-xs font-mono"
                        :class="ev.event_type === 'error' ? 'text-red-700' : 'text-gray-600'">
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
        <!-- Tool header -->
        <div class="p-5 border-b border-border shrink-0">
          <h3 class="text-base font-semibold">{{ selectedTool.name }}</h3>
          <p class="text-sm text-muted-foreground mt-1.5 leading-relaxed">{{ selectedTool.description }}</p>
        </div>

        <div class="flex-1 overflow-auto p-5 space-y-5">

          <!-- Arguments: form mode -->
          <div v-if="hasSchema && !showRawJson" class="space-y-4">
            <div class="flex items-center justify-between">
              <p class="text-xs font-semibold text-muted-foreground uppercase tracking-wide">{{ i18n.t('tv_arguments') }}</p>
              <button @click="showRawJson = true" class="text-xs text-muted-foreground hover:text-foreground underline underline-offset-2">{{ i18n.t('tv_edit_json') }}</button>
            </div>

            <div v-for="prop in schemaProperties" :key="prop.key" class="space-y-1.5">
              <div class="flex items-center gap-2 flex-wrap">
                <span class="text-sm font-medium">{{ prop.key }}</span>
                <span v-if="prop.required" class="text-xs text-red-500 font-medium px-1.5 py-0.5 bg-red-50 rounded">{{ i18n.t('tv_required') }}</span>
                <span class="text-xs text-muted-foreground font-mono bg-gray-100 px-1.5 py-0.5 rounded">{{ prop.type }}</span>
              </div>
              <p v-if="prop.description" class="text-xs text-muted-foreground leading-relaxed">{{ prop.description }}</p>

              <!-- enum: select -->
              <select v-if="prop.enum?.length" v-model="formValues[prop.key]" @change="syncFormToJson"
                class="w-full px-3 py-2 text-sm bg-input border border-border rounded-md focus:outline-none focus:ring-1 focus:ring-ring">
                <option v-for="opt in prop.enum" :key="opt" :value="opt">{{ opt }}</option>
              </select>
              <!-- boolean: checkbox -->
              <label v-else-if="prop.type === 'boolean'" class="flex items-center gap-2 cursor-pointer">
                <input type="checkbox" v-model="formValues[prop.key]" @change="syncFormToJson" class="w-4 h-4 rounded" />
                <span class="text-sm">{{ formValues[prop.key] ? 'true' : 'false' }}</span>
              </label>
              <!-- number -->
              <input v-else-if="prop.type === 'number' || prop.type === 'integer'" type="number"
                v-model.number="formValues[prop.key]" @input="syncFormToJson"
                class="w-full px-3 py-2 text-sm font-mono bg-input border border-border rounded-md focus:outline-none focus:ring-1 focus:ring-ring" />
              <!-- object / array: textarea JSON -->
              <textarea v-else-if="prop.type === 'object' || prop.type === 'array'"
                v-model="formValues[prop.key] as string" @input="syncFormToJson"
                rows="3"
                class="w-full px-3 py-2 text-xs font-mono bg-input border border-border rounded-md focus:outline-none focus:ring-1 focus:ring-ring resize-y"
                :placeholder="prop.type === 'object' ? '{}' : '[]'" />
              <!-- string: textarea -->
              <textarea v-else
                v-model="formValues[prop.key] as string" @input="syncFormToJson"
                rows="2"
                class="w-full px-3 py-2 text-sm font-mono bg-input border border-border rounded-md focus:outline-none focus:ring-1 focus:ring-ring resize-y"
                :placeholder="prop.description || prop.key" />
            </div>
          </div>

          <!-- Arguments: raw JSON mode -->
          <div v-else>
            <div class="flex items-center justify-between mb-1.5">
              <p class="text-xs font-semibold text-muted-foreground uppercase tracking-wide">{{ i18n.t('tv_arguments_json') }}</p>
              <button v-if="hasSchema" @click="showRawJson = false" class="text-xs text-muted-foreground hover:text-foreground underline underline-offset-2">{{ i18n.t('tv_back_to_form') }}</button>
            </div>
            <textarea v-model="argsInput" rows="10"
              class="w-full p-3 text-xs font-mono bg-input border border-border rounded-md focus:outline-none focus:ring-1 focus:ring-ring resize-none" />
          </div>

          <!-- Execute button -->
          <button @click="execute" :disabled="executing"
            class="w-full py-3 text-sm font-semibold rounded-xl transition-colors disabled:opacity-50"
            :class="executing ? 'bg-gray-200 text-gray-500 cursor-not-allowed' : 'bg-gray-900 text-white hover:bg-gray-700'">
            {{ executing ? i18n.t('tv_executing') : i18n.t('tv_execute') }}
          </button>

          <!-- Result -->
          <div v-if="result || execError">
            <p class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-2">{{ execError ? i18n.t('tv_error') : i18n.t('tv_result') }}</p>
            <pre class="p-3 text-xs font-mono rounded-xl border overflow-auto max-h-80"
              :class="execError ? 'border-destructive/50 text-destructive bg-destructive/10' : 'border-border bg-gray-50'">{{ execError || result }}</pre>
          </div>
        </div>
      </template>

    </div>
  </div>
</template>


<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.25s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
