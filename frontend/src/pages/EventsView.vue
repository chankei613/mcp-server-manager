<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useServersStore } from '@/stores/servers'
import type { ServerEventPayload } from '@/lib/wailsTypes'
import { db } from '../../wailsjs/go/models'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

const route = useRoute()
const router = useRouter()
const store = useServersStore()

const serverID = Number(route.params.id)
const server = store.servers.find(s => s.ID === serverID)
const events = ref<db.MCPEvent[]>([])

const eventTypeColors: Record<string, string> = {
  connected: 'text-green-400',
  disconnected: 'text-gray-400',
  error: 'text-red-400',
  degraded: 'text-yellow-400',
  stderr: 'text-orange-400',
  tool_call: 'text-blue-400',
}

onMounted(async () => {
  events.value = await store.fetchEvents(serverID)

  EventsOn('server:event', (payload: ServerEventPayload) => {
    if (payload.serverID === serverID) {
      const event = db.MCPEvent.createFrom({
        ID: Date.now(),
        CreatedAt: new Date(),
        ServerID: serverID,
        event_type: payload.eventType,
        message: payload.message,
      })
      events.value.unshift(event)
    }
  })
})

onUnmounted(() => {
  EventsOff('server:event')
})

function formatTime(val: any): string {
  if (!val) return ''
  try {
    return new Date(val).toLocaleTimeString()
  } catch {
    return String(val)
  }
}
</script>

<template>
  <div class="flex flex-col h-full">
    <div class="p-4 border-b border-border flex items-center gap-2">
      <button @click="router.back()" class="text-muted-foreground hover:text-foreground text-sm">←</button>
      <div>
        <h2 class="text-sm font-medium">{{ server?.name }} — Events</h2>
        <p class="text-xs text-muted-foreground">Live connection log</p>
      </div>
    </div>
    <div class="flex-1 overflow-auto p-4 font-mono text-xs space-y-1">
      <div v-if="events.length === 0" class="text-muted-foreground">No events yet.</div>
      <div v-for="event in events" :key="event.ID" class="flex gap-3">
        <span class="text-muted-foreground shrink-0">{{ formatTime(event.CreatedAt) }}</span>
        <span :class="eventTypeColors[event.event_type] || 'text-foreground'" class="shrink-0 w-20">
          {{ event.event_type }}
        </span>
        <span class="text-muted-foreground break-all">{{ event.message }}</span>
      </div>
    </div>
  </div>
</template>
