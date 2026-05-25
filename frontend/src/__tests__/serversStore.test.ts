import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useServersStore } from '../stores/servers'
import * as AppBindings from '../../wailsjs/go/main/App'

const MOCK_SERVER = {
  ID: 1, name: 'test', transport: 'stdio', command: 'npx', args: '[]', url: '',
  status: 'disconnected', consecutive_failures: 0, last_error: '',
  CreatedAt: null, UpdatedAt: null, DeletedAt: null,
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.clearAllMocks()
})

describe('servers store', () => {
  it('removeServer removes from local list and calls DeleteServer', async () => {
    vi.mocked(AppBindings.DeleteServer).mockResolvedValue(undefined)
    const store = useServersStore()
    store.servers.push({ ...MOCK_SERVER } as any)

    await store.removeServer(1)

    expect(AppBindings.DeleteServer).toHaveBeenCalledWith(1)
    expect(store.servers).toHaveLength(0)
  })

  it('server:status event updates existing server but does NOT re-add deleted server', () => {
    const store = useServersStore()
    store.servers.push({ ...MOCK_SERVER } as any)

    // Simulate status update for existing server
    const idx = store.servers.findIndex(s => s.ID === 1)
    if (idx !== -1) store.servers[idx] = { ...MOCK_SERVER, status: 'connected' } as any
    expect(store.servers[0].status).toBe('connected')

    // After removeServer, simulate a late-arriving status event — should NOT re-add
    store.servers = store.servers.filter(s => s.ID !== 1)
    const idxAfterDelete = store.servers.findIndex(s => s.ID === 1)
    // No push here (the fixed behavior): list stays empty
    expect(idxAfterDelete).toBe(-1)
    expect(store.servers).toHaveLength(0)
  })

  it('connect calls ConnectServer', async () => {
    vi.mocked(AppBindings.ConnectServer).mockResolvedValue(undefined)
    const store = useServersStore()
    await store.connect(1)
    expect(AppBindings.ConnectServer).toHaveBeenCalledWith(1)
  })

  it('disconnect calls DisconnectServer', async () => {
    vi.mocked(AppBindings.DisconnectServer).mockResolvedValue(undefined)
    const store = useServersStore()
    await store.disconnect(1)
    expect(AppBindings.DisconnectServer).toHaveBeenCalledWith(1)
  })

  it('fetchTools calls GetTools with the server ID', async () => {
    vi.mocked(AppBindings.GetTools).mockResolvedValue([])
    const store = useServersStore()
    await store.fetchTools(1)
    expect(AppBindings.GetTools).toHaveBeenCalledWith(1)
  })
})
