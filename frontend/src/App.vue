<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterView, RouterLink, useRoute } from 'vue-router'
import { useServersStore } from '@/stores/servers'
import { useI18nStore } from '@/stores/i18n'
import { EventsOn } from '../wailsjs/runtime/runtime'

const store = useServersStore()
const i18n = useI18nStore()
const route = useRoute()

const updateInfo = ref<{ version: string; release_url: string } | null>(null)

onMounted(() => {
  store.fetchServers()
  store.subscribeToEvents()
  EventsOn('update:available', (info: { version: string; release_url: string }) => {
    updateInfo.value = info
  })
})

function isActive(prefix: string) {
  return route.path.startsWith(prefix)
}
</script>

<template>
  <div class="flex h-screen bg-background text-foreground overflow-hidden">
    <!-- Sidebar -->
    <aside class="w-56 border-r border-border flex flex-col shrink-0 bg-background">
      <div class="p-4 border-b border-border">
        <h1 class="text-sm font-semibold tracking-tight text-foreground">MCP Server Manager</h1>
        <p class="text-xs text-muted-foreground mt-0.5">Visual Client for MCP Servers</p>
      </div>

      <nav class="flex-1 p-2 space-y-0.5">
        <RouterLink
          to="/servers"
          class="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm transition-colors"
          :class="isActive('/servers')
            ? 'bg-gray-900 text-white'
            : 'text-gray-500 hover:bg-gray-100 hover:text-gray-900'"
        >
          <svg class="w-4 h-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="2" y="3" width="20" height="14" rx="2" />
            <path d="M8 21h8M12 17v4" />
          </svg>
          <span>{{ i18n.t('nav_servers') }}</span>
          <span class="ml-auto text-xs opacity-60">{{ store.servers.length }}</span>
        </RouterLink>

        <RouterLink
          to="/import"
          class="flex items-center gap-2.5 px-3 py-2 rounded-md text-sm transition-colors"
          :class="isActive('/import')
            ? 'bg-gray-900 text-white'
            : 'text-gray-500 hover:bg-gray-100 hover:text-gray-900'"
        >
          <svg class="w-4 h-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
            <polyline points="7 10 12 15 17 10" />
            <line x1="12" y1="15" x2="12" y2="3" />
          </svg>
          <span>{{ i18n.t('nav_import') }}</span>
        </RouterLink>
      </nav>

      <div class="p-3 border-t border-border flex items-center justify-between">
        <p class="text-xs text-muted-foreground">v0.1.0</p>
        <button
          @click="i18n.setLang(i18n.lang === 'ja' ? 'en' : 'ja')"
          class="text-xs px-2 py-1 border border-border rounded hover:bg-accent transition-colors text-muted-foreground font-medium"
          :title="i18n.lang === 'ja' ? 'Switch to English' : '日本語に切り替え'"
        >{{ i18n.t('lang_toggle') }}</button>
      </div>
    </aside>

    <!-- Main content -->
    <main class="flex-1 overflow-auto bg-gray-50/50 flex flex-col">
      <!-- Update banner -->
      <div v-if="updateInfo" class="shrink-0 flex items-center justify-between px-4 py-2 bg-indigo-600 text-white text-xs">
        <span>
          {{ i18n.lang === 'ja' ? `v${updateInfo.version} が利用可能です` : `v${updateInfo.version} is available` }}
        </span>
        <div class="flex items-center gap-3">
          <a :href="updateInfo.release_url" target="_blank"
            class="underline underline-offset-2 hover:opacity-80 font-medium">
            {{ i18n.lang === 'ja' ? 'ダウンロード' : 'Download' }}
          </a>
          <button @click="updateInfo = null" class="hover:opacity-70">✕</button>
        </div>
      </div>
      <RouterView class="flex-1" />
    </main>
  </div>
</template>
