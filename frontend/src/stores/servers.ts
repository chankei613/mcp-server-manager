import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ServerEventPayload } from '@/lib/wailsTypes'
import { db, mcp, importer } from '../../wailsjs/go/models'
import {
  GetServers, AddServer, UpdateServer, DeleteServer,
  ConnectServer, DisconnectServer,
  GetTools, CallTool,
  GetEvents,
  GetClaudeDesktopConfigPath, GetLocalClaudeConfigPath, GetCursorConfigPath, GetWindsurfConfigPath,
  ImportClaudeDesktopConfig, OpenFilePicker,
} from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'

export const useServersStore = defineStore('servers', () => {
  const servers = ref<db.MCPServer[]>([])
  const toolsCache = ref<Record<number, mcp.Tool[]>>({})
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchServers() {
    loading.value = true
    try {
      servers.value = await GetServers()
    } catch (e) {
      error.value = String(e)
    } finally {
      loading.value = false
    }
  }

  async function addServer(params: {
    name: string
    transport: string
    command: string
    args: string
    url: string
  }) {
    const server = await AddServer(
      params.name, params.transport, params.command, params.args, params.url
    )
    servers.value.push(server)
    return server
  }

  async function updateServer(id: number, params: {
    name: string, command: string, args: string, url: string
  }) {
    const updated = await UpdateServer(id, params.name, params.command, params.args, params.url)
    const idx = servers.value.findIndex(s => s.ID === id)
    if (idx !== -1) servers.value[idx] = updated
  }

  async function removeServer(id: number) {
    await DeleteServer(id)
    servers.value = servers.value.filter(s => s.ID !== id)
  }

  async function connect(id: number) {
    const idx = servers.value.findIndex(s => s.ID === id)
    if (idx !== -1) servers.value[idx].status = 'connecting'
    await ConnectServer(id)
  }

  async function disconnect(id: number) {
    await DisconnectServer(id)
  }

  async function fetchTools(serverID: number): Promise<mcp.Tool[]> {
    const tools = await GetTools(serverID)
    toolsCache.value[serverID] = tools
    return tools
  }

  async function callTool(serverID: number, toolName: string, args: string): Promise<mcp.ToolCallResult> {
    return CallTool(serverID, toolName, args)
  }

  async function fetchEvents(serverID: number): Promise<db.MCPEvent[]> {
    return GetEvents(serverID)
  }

  async function importClaudeDesktop(): Promise<importer.ImportResult> {
    const configPath = await GetClaudeDesktopConfigPath()
    const result = await ImportClaudeDesktopConfig(configPath)
    await fetchServers()
    return result
  }

  async function importFromPath(path: string): Promise<importer.ImportResult> {
    const result = await ImportClaudeDesktopConfig(path)
    await fetchServers()
    return result
  }

  async function getLocalClaudePath(): Promise<string> {
    return GetLocalClaudeConfigPath()
  }

  async function getCursorConfigPath(): Promise<string> {
    return GetCursorConfigPath()
  }

  async function getWindsurfConfigPath(): Promise<string> {
    return GetWindsurfConfigPath()
  }

  async function pickFile(): Promise<string> {
    return OpenFilePicker()
  }

  function subscribeToEvents() {
    EventsOn('server:status', (updated: db.MCPServer) => {
      const idx = servers.value.findIndex(s => s.ID === updated.ID)
      if (idx !== -1) servers.value[idx] = updated
      else servers.value.push(updated)
    })

    EventsOn('server:event', (_payload: ServerEventPayload) => {
      // handled in EventsView
    })
  }

  return {
    servers, toolsCache, loading, error,
    fetchServers, addServer, updateServer, removeServer,
    connect, disconnect, fetchTools, callTool, fetchEvents,
    importClaudeDesktop, importFromPath, getLocalClaudePath, getCursorConfigPath, getWindsurfConfigPath, pickFile,
    subscribeToEvents,
  }
})
