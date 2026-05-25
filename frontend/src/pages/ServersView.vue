<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useServersStore } from '@/stores/servers'
import { useI18nStore } from '@/stores/i18n'

const store = useServersStore()
const router = useRouter()
const i18n = useI18nStore()

onMounted(() => {
  store.fetchServers()
})

const showAddForm = ref(false)
const form = ref({ name: '', transport: 'stdio', command: '', args: '[]', url: '' })
const connecting = ref<number | null>(null)
const pendingDeleteId = ref<number | null>(null)

const statusColors: Record<string, string> = {
  connected: 'bg-green-500',
  disconnected: 'bg-gray-500',
  error: 'bg-red-500',
  degraded: 'bg-yellow-500',
  connecting: 'bg-blue-500 animate-pulse',
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

function requestDelete(id: number) {
  pendingDeleteId.value = id
}

async function confirmDelete(id: number) {
  pendingDeleteId.value = null
  await store.removeServer(id)
}
</script>

<template>
  <div class="p-6">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-lg font-semibold">{{ i18n.t('sv_title') }}</h2>
        <p class="text-sm text-muted-foreground">{{ i18n.t('sv_count', store.servers.length) }}</p>
      </div>
    </div>

    <!-- Add form -->
    <div v-if="showAddForm" class="mb-6 p-4 border border-border rounded-lg space-y-3">
      <h3 class="text-sm font-medium">{{ i18n.t('sv_form_title') }}</h3>
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="text-xs text-muted-foreground">{{ i18n.t('sv_form_name') }}</label>
          <input v-model="form.name" placeholder="my-mcp-server"
            class="mt-1 w-full px-3 py-1.5 text-sm bg-input border border-border rounded-md focus:outline-none focus:ring-1 focus:ring-ring" />
        </div>
        <div>
          <label class="text-xs text-muted-foreground">{{ i18n.t('sv_form_transport') }}</label>
          <select v-model="form.transport"
            class="mt-1 w-full px-3 py-1.5 text-sm bg-input border border-border rounded-md focus:outline-none">
            <option value="stdio">stdio</option>
            <option value="http">HTTP</option>
          </select>
        </div>
      </div>
      <template v-if="form.transport === 'stdio'">
        <div>
          <label class="text-xs text-muted-foreground">{{ i18n.t('sv_form_command') }}</label>
          <input v-model="form.command" placeholder="npx"
            class="mt-1 w-full px-3 py-1.5 text-sm bg-input border border-border rounded-md focus:outline-none" />
        </div>
        <div>
          <label class="text-xs text-muted-foreground">{{ i18n.t('sv_form_args') }}</label>
          <input v-model="form.args" placeholder="[&quot;-y&quot;, &quot;@modelcontextprotocol/server-filesystem&quot;, &quot;/path&quot;]"
            class="mt-1 w-full px-3 py-1.5 text-sm bg-input border border-border rounded-md focus:outline-none font-mono" />
        </div>
      </template>
      <template v-else>
        <div>
          <label class="text-xs text-muted-foreground">{{ i18n.t('sv_form_url') }}</label>
          <input v-model="form.url" placeholder="http://localhost:8080"
            class="mt-1 w-full px-3 py-1.5 text-sm bg-input border border-border rounded-md focus:outline-none" />
        </div>
      </template>
      <div class="flex gap-2">
        <button @click="handleAdd"
          class="px-3 py-1.5 text-sm bg-primary text-primary-foreground rounded-md hover:opacity-90">{{ i18n.t('sv_form_add') }}</button>
        <button @click="showAddForm = false"
          class="px-3 py-1.5 text-sm border border-border rounded-md hover:bg-accent">{{ i18n.t('sv_form_cancel') }}</button>
      </div>
    </div>

    <!-- Empty state / Onboarding -->
    <div v-if="store.servers.length === 0" class="max-w-xl mx-auto py-12">
      <div class="text-center mb-8">
        <div class="w-14 h-14 bg-primary/10 rounded-full flex items-center justify-center mx-auto mb-4">
          <svg class="w-7 h-7 text-primary" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <rect x="2" y="3" width="20" height="14" rx="2" />
            <path d="M8 21h8M12 17v4" />
          </svg>
        </div>
        <h3 class="text-base font-semibold text-foreground">{{ i18n.t('sv_empty_title') }}</h3>
        <p class="text-sm text-muted-foreground mt-1">{{ i18n.t('sv_empty_desc') }}</p>
      </div>

      <div class="space-y-3 mb-8">
        <div class="flex items-start gap-3 p-4 border border-border rounded-lg bg-white">
          <span class="w-6 h-6 rounded-full bg-primary text-primary-foreground text-xs font-bold flex items-center justify-center shrink-0 mt-0.5">1</span>
          <div>
            <p class="text-sm font-medium">{{ i18n.t('sv_step1_title') }}</p>
            <p class="text-xs text-muted-foreground mt-0.5">{{ i18n.t('sv_step1_desc') }}</p>
            <button
              @click="$router.push('/import')"
              class="mt-2 text-xs text-primary font-medium hover:opacity-70 transition-opacity"
            >{{ i18n.t('sv_step1_link') }}</button>
          </div>
        </div>

        <div class="flex items-start gap-3 p-4 border border-border rounded-lg bg-white">
          <span class="w-6 h-6 rounded-full bg-primary text-primary-foreground text-xs font-bold flex items-center justify-center shrink-0 mt-0.5">2</span>
          <div>
            <p class="text-sm font-medium">{{ i18n.t('sv_step2_title') }}</p>
            <p class="text-xs text-muted-foreground mt-0.5">{{ i18n.t('sv_step2_desc') }}</p>
            <button
              @click="showAddForm = true"
              class="mt-2 text-xs text-primary font-medium hover:opacity-70 transition-opacity"
            >→ {{ i18n.t('sv_add_btn') }}</button>
          </div>
        </div>

        <div class="flex items-start gap-3 p-4 border border-border rounded-lg bg-white">
          <span class="w-6 h-6 rounded-full bg-primary text-primary-foreground text-xs font-bold flex items-center justify-center shrink-0 mt-0.5">3</span>
          <div>
            <p class="text-sm font-medium">{{ i18n.t('sv_step3_title') }}</p>
            <p class="text-xs text-muted-foreground mt-0.5">{{ i18n.t('sv_step3_desc') }}</p>
          </div>
        </div>
      </div>

      <div class="flex gap-3 justify-center">
        <button
          @click="$router.push('/import')"
          class="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
        >{{ i18n.t('sv_import_config') }}</button>
        <button
          @click="showAddForm = true"
          class="px-4 py-2 text-sm border border-border rounded-lg hover:bg-accent transition-colors"
        >{{ i18n.t('sv_add_manual') }}</button>
      </div>
    </div>

    <!-- Server list -->
    <div class="space-y-2">
      <div
        v-for="server in store.servers"
        :key="server.ID"
        @click="router.push(`/servers/${server.ID}/tools`)"
        class="flex items-center gap-3 p-4 border border-border rounded-lg hover:bg-accent/20 transition-colors cursor-pointer"
      >
        <span class="w-2 h-2 rounded-full shrink-0" :class="statusColors[server.status]" />
        <div class="flex-1 min-w-0">
          <p class="text-sm font-medium">{{ server.name }}</p>
          <p v-if="server.status === 'error' && server.last_error" class="text-xs text-red-600 truncate mt-0.5">{{ server.last_error }}</p>
        </div>
        <div class="flex gap-2 shrink-0">
          <button
            v-if="server.status === 'connecting' || connecting === server.ID"
            disabled
            @click.stop
            class="text-xs px-3 py-1.5 rounded-md bg-blue-50 text-blue-600 border border-blue-200 cursor-not-allowed"
          >{{ i18n.t('sv_connecting') }}</button>
          <button
            v-else-if="server.status === 'connected'"
            @click.stop="handleDisconnect(server.ID)"
            class="text-xs px-3 py-1.5 rounded-md bg-green-50 text-green-700 border border-green-300 hover:bg-green-100 transition-colors"
          >{{ i18n.t('sv_connected') }}</button>
          <button
            v-else-if="server.status === 'error'"
            @click.stop="handleConnect(server.ID)"
            :disabled="connecting === server.ID"
            class="text-xs px-3 py-1.5 rounded-md bg-red-50 text-red-700 border border-red-300 hover:bg-red-100 transition-colors disabled:opacity-50"
          >{{ i18n.t('sv_retry') }}</button>
          <button
            v-else
            @click.stop="handleConnect(server.ID)"
            :disabled="connecting === server.ID"
            class="text-xs px-3 py-1.5 rounded-md border border-border hover:bg-accent transition-colors disabled:opacity-50"
          >{{ i18n.t('sv_connect') }}</button>

          <!-- Delete: 2-step inline confirm -->
          <button
            v-if="pendingDeleteId !== server.ID"
            @click.stop="requestDelete(server.ID)"
            class="text-xs px-2 py-1.5 border border-red-400 text-red-500 rounded-md hover:bg-red-50 transition-colors"
          >{{ i18n.t('sv_delete') }}</button>
          <div v-else class="flex gap-1" @click.stop>
            <button
              @click.stop="confirmDelete(server.ID)"
              class="text-xs px-2 py-1.5 bg-red-600 text-white rounded-md hover:bg-red-700 transition-colors"
            >{{ i18n.lang === 'ja' ? '確認して削除' : 'Delete' }}</button>
            <button
              @click.stop="pendingDeleteId = null"
              class="text-xs px-2 py-1.5 border border-gray-300 text-gray-500 rounded-md hover:bg-gray-100 transition-colors"
            >✕</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
