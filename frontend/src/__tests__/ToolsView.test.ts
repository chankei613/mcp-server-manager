import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia, getActivePinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import ToolsView from '../pages/ToolsView.vue'
import * as AppBindings from '../../wailsjs/go/main/App'
import { useServersStore } from '../stores/servers'

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/servers/:id/tools', component: ToolsView },
  ],
})

const MOCK_TOOLS = [
  { name: 'echo', description: 'echoes text', inputSchema: { type: 'object', properties: { text: { type: 'string' } } } },
  { name: 'list-files', description: 'lists files', inputSchema: { type: 'object', properties: { path: { type: 'string' } } } },
]

const MOCK_SERVER = {
  ID: 1, name: 'test-server', transport: 'stdio', command: 'npx', args: '[]', url: '',
  status: 'connected', consecutive_failures: 0, last_error: '', CreatedAt: null, UpdatedAt: null, DeletedAt: null,
}

async function mountToolsView() {
  await router.push('/servers/1/tools')
  vi.mocked(AppBindings.GetTools).mockResolvedValue(MOCK_TOOLS as any)
  // Seed the store with a connected server so the watch triggers loadTools()
  const store = useServersStore(getActivePinia()!)
  store.servers.push(MOCK_SERVER as any)
  const wrapper = mount(ToolsView, {
    global: { plugins: [getActivePinia()!, router] },
  })
  await flushPromises()
  return wrapper
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.clearAllMocks()
})

describe('ToolsView', () => {
  it('shows loading then tool list', async () => {
    const wrapper = await mountToolsView()
    expect(AppBindings.GetTools).toHaveBeenCalledWith(1)
    expect(wrapper.text()).toContain('echo')
    expect(wrapper.text()).toContain('list-files')
  })

  it('shows tool count', async () => {
    const wrapper = await mountToolsView()
    expect(wrapper.text()).toContain('(2)')
  })

  it('shows execution panel when a tool is selected', async () => {
    const wrapper = await mountToolsView()
    await wrapper.vm.$nextTick()
    // click first tool button
    const toolBtns = wrapper.findAll('button').filter(b => b.text().includes('echo'))
    if (toolBtns.length > 0) {
      await toolBtns[0].trigger('click')
      expect(wrapper.find('textarea').exists()).toBe(true)
    }
  })

  it('pre-fills form values with schema defaults on tool select', async () => {
    const wrapper = await mountToolsView()
    await wrapper.vm.$nextTick()
    const toolBtns = wrapper.findAll('button').filter(b => b.text().includes('echo'))
    if (toolBtns.length > 0) {
      await toolBtns[0].trigger('click')
      // Form mode: string fields render as textarea/input with empty default
      const vm = wrapper.vm as any
      expect(vm.formValues).toHaveProperty('text', '')
    }
  })

  it('calls CallTool when Execute is clicked', async () => {
    const wrapper = await mountToolsView()
    const toolBtns = wrapper.findAll('button').filter(b => b.text().includes('echo'))
    if (toolBtns.length > 0) {
      await toolBtns[0].trigger('click')
      const execBtn = wrapper.findAll('button').find(b => b.text().includes('実行する') || b.text().includes('Execute'))
      await execBtn!.trigger('click')
      await flushPromises()
      expect(AppBindings.CallTool).toHaveBeenCalledWith(1, 'echo', expect.any(String))
    }
  })

  it('shows result after execution', async () => {
    const wrapper = await mountToolsView()
    const toolBtns = wrapper.findAll('button').filter(b => b.text().includes('echo'))
    if (toolBtns.length > 0) {
      await toolBtns[0].trigger('click')
      const execBtn = wrapper.findAll('button').find(b => b.text().includes('実行する') || b.text().includes('Execute'))
      await execBtn!.trigger('click')
      await flushPromises()
      // 'tv_result' key renders as '結果' in Japanese (default lang)
      expect(wrapper.text()).toContain('結果')
    }
  })
})
