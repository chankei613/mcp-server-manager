package mcp

import "context"

// Transport はMCPサーバーとの通信手段を抽象化するインターフェース
type Transport interface {
	Start(ctx context.Context) error
	Stop() error
	Call(ctx context.Context, method string, params interface{}) (*Response, error)
	Notify(method string, params interface{}) error
	IsRunning() bool
}
