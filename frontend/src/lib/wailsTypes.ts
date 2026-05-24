// Wails生成型を再エクスポートして便利な型エイリアスを定義
export type { db, mcp, importer } from '../../wailsjs/go/models'

// Wails生成型のショートカット（db.MCPServerの代わりに使用可）
export type ServerStatus = 'connected' | 'disconnected' | 'error' | 'degraded' | 'connecting'

// Wailsイベントのペイロード型
export interface ServerEventPayload {
  serverID: number
  eventType: string
  message: string
}
