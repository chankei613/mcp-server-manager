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
</script>

<template>
  <div class="flex h-screen bg-background text-foreground overflow-hidden">
    <!-- Sidebar -->
    <aside class="w-56 border-r border-border flex flex-col shrink-0">
      <div class="p-4 border-b border-border">
        <h1 class="text-sm font-semibold tracking-tight">MCP Server Manager</h1>
        <p class="text-xs text-muted-foreground mt-0.5">Postman for MCP</p>
      </div>
      <nav class="flex-1 p-2 space-y-1">
        <RouterLink
          to="/servers"
          class="flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors hover:bg-accent hover:text-accent-foreground"
          :class="route.path.startsWith('/servers') ? 'bg-accent text-accent-foreground' : 'text-muted-foreground'"
        >
          <span>Servers</span>
        </RouterLink>
      </nav>
      <div class="p-3 border-t border-border">
        <p class="text-xs text-muted-foreground">v0.1.0</p>
      </div>
    </aside>

    <!-- Main content -->
    <main class="flex-1 overflow-auto">
      <RouterView />
    </main>
  </div>
</template>
