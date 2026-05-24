package db

import (
	"time"

	"gorm.io/gorm"
)

// TransportType はMCPサーバーの接続方式
type TransportType string

const (
	TransportStdio TransportType = "stdio"
	TransportHTTP  TransportType = "http"
)

// ServerStatus はMCPサーバーの接続状態
type ServerStatus string

const (
	StatusConnected    ServerStatus = "connected"
	StatusDisconnected ServerStatus = "disconnected"
	StatusError        ServerStatus = "error"
	StatusDegraded     ServerStatus = "degraded"
	StatusConnecting   ServerStatus = "connecting"
)

// MCPServer はMCPサーバーの設定と状態
type MCPServer struct {
	gorm.Model
	Name      string        `json:"name" gorm:"uniqueIndex;not null"`
	Transport TransportType `json:"transport" gorm:"not null"`

	// stdio transport
	Command string `json:"command"`
	Args    string `json:"args"` // JSON配列文字列として保存

	// HTTP transport
	URL string `json:"url"`

	// 状態
	Status             ServerStatus `json:"status" gorm:"default:disconnected"`
	ConsecutiveFailures int         `json:"consecutive_failures" gorm:"default:0"`
	LastConnectedAt    *time.Time   `json:"last_connected_at"`
	LastErrorAt        *time.Time   `json:"last_error_at"`
	LastError          string       `json:"last_error"`

	// リレーション
	Events    []MCPEvent     `json:"events,omitempty" gorm:"foreignKey:ServerID"`
	Executions []ToolExecution `json:"executions,omitempty" gorm:"foreignKey:ServerID"`
}

// MCPEvent はサーバーの接続ログ・エラー
type MCPEvent struct {
	gorm.Model
	ServerID  uint   `json:"server_id" gorm:"index;not null"`
	EventType string `json:"event_type"` // connected, disconnected, error, tool_call
	Message   string `json:"message"`
}

// ToolExecution はツール実行履歴
type ToolExecution struct {
	gorm.Model
	ServerID   uint   `json:"server_id" gorm:"index;not null"`
	ToolName   string `json:"tool_name" gorm:"not null"`
	Parameters string `json:"parameters"` // JSON文字列
	Result     string `json:"result"`     // JSON文字列
	Error      string `json:"error"`
	DurationMs int64  `json:"duration_ms"`
}
