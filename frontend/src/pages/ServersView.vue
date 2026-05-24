<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useServersStore } from '@/stores/servers'

const store = useServersStore()
const router = useRouter()

const showAddForm = ref(false)
const form = ref({ name: '', transport: 'stdio', command: '', args: '[]', url: '' })
const connecting = ref<number | null>(null)
const importing = ref(false)
const importResult = ref<{ imported: string[], skipped: string[], errors: string[] } | null>(null)
const importError = ref<string | null>(null)

const statusColors: Record<string, string> = {
  connected: 'bg-green-500',
  disconnected: 'bg-gray-500',
  error: 'bg-red-500',
  degraded: 'bg-yellow-500',
  connecting: 'bg-blue-500 animate-pulse',
}

const statusLabel: Record<string, string> = {
  connected: 'Connected',
  disconnected: 'Disconnected',
  error: 'Error',
  degraded: 'Degraded',
  connecting: 'Connecting…',
}

async function handleAdd() {
  await store.addServer(form.value)
  showAddForm.value = false
  form.value = { name: '', transport: 'stdio', command: '', args: '[]', url: '' }
}

async function handleConnect(id: number) {
  connecting.value = id
  try {
    await store.connect(id)
  } catch (e) {
    console.error(e)
  } finally {
    connecting.value = null
  }
}

async function handleDisconnect(id: number) {
  await store.disconnect(id)
}

async function handleDelete(id: number) {
  if (confirm('Delete this server?')) await store.removeServer(id)
}

function openTools(id: number) {
  router.push(`/servers/${id}/tools`)
}

async function handleImport() {
  importing.value = true
  importResult.value = null
  importError.value = null
  try {
    importResult.value = await store.importClaudeDesktop()
  } catch (e) {
    importError.value = String(e)
  } finally {
    importing.value = false
  }
}
</script>

<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-lg font-semibold">MCP Servers</h2>
        <p class="text-sm text-muted-foreground">{{ store.servers.length }} server(s) configured</p>
      </div>
      <div class="flex gap-2">
        <button
          @click="handleImport"
          :disabled="importing"
          class="px-3 py-1.5 text-sm border border-border rounded-md hover:bg-accent disabled:opacity-50 transition-colors"
        >
          {{ importing ? 'Importing…' : '↓ Import Claude Desktop' }}
        </button>
        <button
          @click="showAddForm = !showAddForm"
          class="px-3 py-1.5 text-sm bg-primary text-primary-foreground rounded-md hover:opacity-90 transition-opacity"
        >
          + Add Server
        </button>
      </div>
    </div>

    <!-- Import result banner -->
    <div v-if="importResult" class="mb-4 p-3 border border-border rounded-lg text-xs space-y-1">
      <p class="font-medium">Import complete</p>
      <p v-if="importResult.imported?.length" class="text-green-400">
        ✓ Imported: {{ importResult.imported.join(', ') }}
      </p>
      <p v-if="importResult.skipped?.length" class="text-muted-foreground">
        — Skipped (already exist): {{ importResult.skipped.join(', ') }}
      </p>
      <p v-for="err in importResult.errors" :key="err" class="text-destructive">✗ {{ err }}</p>
      <button @click="importResult = null" class="text-muted-foreground hover:text-foreground mt-1">Dismiss</button>
    </div>
    <div v-if="importError" class="mb-4 p-3 border border-destructive/50 rounded-lg text-xs text-destructive">
      {{ importError }}
      <button @click="importError = null" class="ml-2 hover:opacity-70">✕</button>
    </div>

    <!-- Add form -->
    <div v-if="showAddForm" class="mb-6 p-4 border border-border rounded-lg space-y-3">
      <h3 class="text-sm font-medium">New Server</h3>
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="text-xs text-muted-foreground">Name</label>
          <input v-model="form.name" placeholder="my-mcp-server"
            class="mt-1 w-full px-3 py-1.5 text-sm bg-input border border-border rounded-md focus:outline-none focus:ring-1 focus:ring-ring" />
        </div>
        <div>
          <label class="text-xs text-muted-foreground">Transport</label>
          <select v-model="form.transport"
            class="mt-1 w-full px-3 py-1.5 text-sm bg-input border border-border rounded-md focus:outline-none">
            <option value="stdio">stdio</option>
            <option value="http">HTTP</option>
          </select>
        </div>
      </div>
      <template v-if="form.transport === 'stdio'">
        <div>
          <label class="text-xs text-muted-foreground">Command</label>
          <input v-model="form.command" placeholder="npx"
            class="mt-1 w-full px-3 py-1.5 text-sm bg-input border border-border rounded-md focus:outline-none" />
        </div>
        <div>
          <label class="text-xs text-muted-foreground">Args (JSON array)</label>
          <input v-model="form.args" placeholder="[&quot;-y&quot;, &quot;@modelcontextprotocol/server-filesystem&quot;, &quot;/path&quot;]"
            class="mt-1 w-full px-3 py-1.5 text-sm bg-input border border-border rounded-md focus:outline-none font-mono" />
        </div>
      </template>
      <template v-else>
        <div>
          <label class="text-xs text-muted-foreground">URL</label>
          <input v-model="form.url" placeholder="http://localhost:8080"
            class="mt-1 w-full px-3 py-1.5 text-sm bg-input border border-border rounded-md focus:outline-none" />
        </div>
      </template>
      <div class="flex gap-2">
        <button @click="handleAdd"
          class="px-3 py-1.5 text-sm bg-primary text-primary-foreground rounded-md hover:opacity-90">Add</button>
        <button @click="showAddForm = false"
          class="px-3 py-1.5 text-sm border border-border rounded-md hover:bg-accent">Cancel</button>
      </div>
    </div>

    <!-- Empty state -->
    <div v-if="store.servers.length === 0" class="text-center py-16 text-muted-foreground">
      <p class="text-sm">No servers configured.</p>
      <p class="text-xs mt-1">Click "+ Add Server" to add an MCP server, or import from Claude Desktop.</p>
    </div>

    <!-- Server list -->
    <div class="space-y-2">
      <div
        v-for="server in store.servers"
        :key="server.ID"
        class="flex items-center gap-3 p-4 border border-border rounded-lg hover:bg-accent/20 transition-colors"
      >
        <span class="w-2 h-2 rounded-full shrink-0" :class="statusColors[server.status]" />
        <div class="flex-1 min-w-0">
          <p class="text-sm font-medium truncate">{{ server.name }}</p>
          <p class="text-xs text-muted-foreground truncate">
            {{ server.transport }} · {{ server.command || server.url }}
          </p>
        </div>
        <span class="text-xs px-2 py-0.5 rounded-full border border-border text-muted-foreground">
          {{ statusLabel[server.status] || server.status }}
        </span>
        <div class="flex gap-2 shrink-0">
          <button
            v-if="server.status !== 'connected'"
            @click="handleConnect(server.ID)"
            :disabled="connecting === server.ID"
            class="text-xs px-2 py-1 border border-border rounded hover:bg-accent disabled:opacity-50"
          >Connect</button>
          <button
            v-else
            @click="openTools(server.ID)"
            class="text-xs px-2 py-1 border border-green-600 text-green-400 rounded hover:bg-green-900/20"
          >Browse Tools</button>
          <button
            v-if="server.status === 'connected'"
            @click="handleDisconnect(server.ID)"
            class="text-xs px-2 py-1 border border-border rounded hover:bg-accent"
          >Disconnect</button>
          <button
            @click="handleDelete(server.ID)"
            class="text-xs px-2 py-1 border border-destructive/50 text-destructive rounded hover:bg-destructive/10"
          >Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>
