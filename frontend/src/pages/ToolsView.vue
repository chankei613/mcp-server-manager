<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
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
const loading = ref(true)

const argsInput = ref('{}')
const executing = ref(false)
const result = ref<string | null>(null)
const execError = ref<string | null>(null)

onMounted(async () => {
  try {
    tools.value = await store.fetchTools(serverID)
  } catch (e) {
    execError.value = String(e)
  } finally {
    loading.value = false
  }
})

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
    <!-- Tool list panel -->
    <div class="w-72 border-r border-border flex flex-col">
      <div class="p-4 border-b border-border flex items-center gap-2">
        <button @click="router.back()" class="text-muted-foreground hover:text-foreground text-sm">←</button>
        <div>
          <h2 class="text-sm font-medium">{{ server?.name }}</h2>
          <p class="text-xs text-muted-foreground">{{ tools.length }} tools</p>
        </div>
      </div>
      <div class="flex-1 overflow-auto p-2">
        <div v-if="loading" class="p-4 text-sm text-muted-foreground">Loading tools…</div>
        <div v-else-if="tools.length === 0" class="p-4 text-sm text-muted-foreground">No tools found.</div>
        <button
          v-for="tool in tools"
          :key="tool.name"
          @click="selectTool(tool)"
          class="w-full text-left px-3 py-2 rounded-md text-sm hover:bg-accent transition-colors mb-0.5"
          :class="selectedTool?.name === tool.name ? 'bg-accent text-accent-foreground' : 'text-muted-foreground'"
        >
          <p class="font-medium text-foreground truncate">{{ tool.name }}</p>
          <p class="text-xs truncate">{{ tool.description }}</p>
        </button>
      </div>
    </div>

    <!-- Execution panel -->
    <div class="flex-1 flex flex-col">
      <div v-if="!selectedTool" class="flex-1 flex items-center justify-center text-muted-foreground">
        <p class="text-sm">Select a tool to execute it</p>
      </div>
      <template v-else>
        <div class="p-4 border-b border-border">
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
          >
            {{ executing ? 'Executing…' : '▶ Execute' }}
          </button>
          <div v-if="result || execError">
            <label class="text-xs text-muted-foreground mb-1 block">
              {{ execError ? 'Error' : 'Result' }}
            </label>
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
