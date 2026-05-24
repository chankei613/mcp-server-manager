import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import ServersView from '../pages/ServersView.vue'
import * as AppBindings from '../../wailsjs/go/main/App'

const router = createRouter({
  history: createMemoryHistory(),
  routes: [{ path: '/', component: ServersView }],
})

function mountServersView() {
  return mount(ServersView, {
    global: {
      plugins: [createPinia(), router],
    },
  })
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.clearAllMocks()
})

describe('ServersView', () => {
  it('renders the page title', () => {
    const wrapper = mountServersView()
    expect(wrapper.text()).toContain('MCP Servers')
  })

  it('shows onboarding when no servers', () => {
    const wrapper = mountServersView()
    expect(wrapper.text()).toContain('MCP サーバーがまだ登録されていません')
  })

  it('shows add form when "+ Add Server" is clicked', async () => {
    const wrapper = mountServersView()
    await wrapper.find('button[class*="bg-primary"]').trigger('click')
    expect(wrapper.text()).toContain('新しいサーバー')
    expect(wrapper.find('input').exists()).toBe(true)
  })

  it('hides add form when Cancel is clicked', async () => {
    const wrapper = mountServersView()
    await wrapper.find('button[class*="bg-primary"]').trigger('click')
    const cancelBtn = wrapper.findAll('button').find(b => b.text() === 'キャンセル')
    await cancelBtn!.trigger('click')
    expect(wrapper.text()).not.toContain('新しいサーバー')
  })

  it('calls AddServer when form is submitted', async () => {
    const wrapper = mountServersView()
    await wrapper.find('button[class*="bg-primary"]').trigger('click')
    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('my-server')
    const addBtn = wrapper.findAll('button').find(b => b.text() === '追加')
    await addBtn!.trigger('click')
    expect(AppBindings.AddServer).toHaveBeenCalledWith('my-server', 'stdio', '', '[]', '')
  })

  it('shows import result after clicking import button', async () => {
    const wrapper = mountServersView()
    const importBtn = wrapper.findAll('button').find(b => b.text().includes('Claude Desktop'))
    await importBtn!.trigger('click')
    // wait for async import
    await vi.runAllTimersAsync().catch(() => {})
    await wrapper.vm.$nextTick()
    await wrapper.vm.$nextTick()
    expect(AppBindings.GetClaudeDesktopConfigPath).toHaveBeenCalled()
    expect(AppBindings.ImportClaudeDesktopConfig).toHaveBeenCalled()
  })

  it('shows server list after servers are loaded', async () => {
    vi.mocked(AppBindings.GetServers).mockResolvedValue([
      { ID: 1, name: 'fs-server', transport: 'stdio', command: 'npx', args: '[]', url: '', status: 'disconnected', consecutive_failures: 0, last_error: '', CreatedAt: null, UpdatedAt: null, DeletedAt: null },
    ] as any)

    const wrapper = mountServersView()
    // trigger store load
    const store = (wrapper.vm as any).store
    await store.fetchServers()
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('fs-server')
  })
})
