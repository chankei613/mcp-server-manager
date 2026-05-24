<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterView, RouterLink, useRoute } from 'vue-router'
import { useServersStore } from '@/stores/servers'

const store = useServersStore()
const route = useRoute()

onMounted(() => {
  store.fetchServers()
  store.subscribeToEvents()
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
        <p class="text-xs text-muted-foreground mt-0.5">Postman for MCP</p>
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
          <span>Servers</span>
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
          <span>Import</span>
        </RouterLink>
      </nav>

      <div class="p-3 border-t border-border">
        <p class="text-xs text-muted-foreground">v0.1.0</p>
      </div>
    </aside>

    <!-- Main content -->
    <main class="flex-1 overflow-auto bg-gray-50/50">
      <RouterView />
    </main>
  </div>
</template>
