import { vi } from 'vitest'

// Wails runtime mock — EventsOn/EventsOff are no-ops in test
vi.mock('../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(),
  EventsOff: vi.fn(),
  EventsEmit: vi.fn(),
}))

// Wails App bindings mock — return empty defaults unless overridden per-test
vi.mock('../../wailsjs/go/main/App', () => ({
  GetServers: vi.fn().mockResolvedValue([]),
  AddServer: vi.fn().mockResolvedValue({ ID: 1, name: 'test', transport: 'stdio', command: 'npx', args: '[]', url: '', status: 'disconnected', consecutive_failures: 0, last_error: '' }),
  UpdateServer: vi.fn().mockResolvedValue({}),
  DeleteServer: vi.fn().mockResolvedValue(undefined),
  ConnectServer: vi.fn().mockResolvedValue(undefined),
  DisconnectServer: vi.fn().mockResolvedValue(undefined),
  GetTools: vi.fn().mockResolvedValue([]),
  CallTool: vi.fn().mockResolvedValue({ content: [{ type: 'text', text: 'ok' }], isError: false }),
  GetEvents: vi.fn().mockResolvedValue([]),
  GetClaudeDesktopConfigPath: vi.fn().mockResolvedValue('/mock/path/claude_desktop_config.json'),
  ImportClaudeDesktopConfig: vi.fn().mockResolvedValue({ imported: ['server-a'], skipped: [], errors: [] }),
  GetAppVersion: vi.fn().mockResolvedValue('0.1.0'),
  GetLocalClaudeConfigPath: vi.fn().mockResolvedValue('/mock/.claude.json'),
  GetCursorConfigPath: vi.fn().mockResolvedValue('/mock/cursor.json'),
  GetWindsurfConfigPath: vi.fn().mockResolvedValue('/mock/windsurf.json'),
  ImportFromPath: vi.fn().mockResolvedValue({ imported: [], skipped: [], errors: [] }),
  PickFile: vi.fn().mockResolvedValue(''),
}))
