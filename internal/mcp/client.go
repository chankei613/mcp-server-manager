package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// Client はMCPサーバーとの高レベルなやり取りを担当する
type Client struct {
	transport Transport
	serverID  uint
	info      *InitializeResult
}

func NewClient(transport Transport, serverID uint) *Client {
	return &Client{transport: transport, serverID: serverID}
}

// Initialize はMCPハンドシェイクを実行する
// MCP仕様: initialize → (server responds) → notifications/initialized
func (c *Client) Initialize(ctx context.Context) (*InitializeResult, error) {
	params := InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities:    ClientCapabilities{},
		ClientInfo: ClientInfo{
			Name:    "mcp-server-manager",
			Version: "0.1.0",
		},
	}

	resp, err := c.transport.Call(ctx, "initialize", params)
	if err != nil {
		return nil, fmt.Errorf("initialize: %w", err)
	}

	var result InitializeResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("parse initialize result: %w", err)
	}

	// ハンドシェイク完了通知を送る（レスポンス不要）
	_ = c.transport.Notify("notifications/initialized", nil)

	c.info = &result
	return &result, nil
}

// ListTools はサーバーから利用可能なツール一覧を取得する
func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	resp, err := c.transport.Call(ctx, "tools/list", nil)
	if err != nil {
		return nil, fmt.Errorf("tools/list: %w", err)
	}

	var result ToolsListResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("parse tools/list result: %w", err)
	}

	return result.Tools, nil
}

// CallTool はツールを実行してその結果を返す
func (c *Client) CallTool(ctx context.Context, name string, arguments json.RawMessage) (*ToolCallResult, error) {
	params := ToolCallParams{
		Name:      name,
		Arguments: arguments,
	}

	resp, err := c.transport.Call(ctx, "tools/call", params)
	if err != nil {
		return nil, fmt.Errorf("tools/call %s: %w", name, err)
	}

	var result ToolCallResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("parse tools/call result: %w", err)
	}

	return &result, nil
}

func (c *Client) ServerInfo() *InitializeResult {
	return c.info
}
