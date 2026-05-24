package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/chankei613/mcp-server-manager/internal/db"
	"github.com/chankei613/mcp-server-manager/internal/health"
	"github.com/chankei613/mcp-server-manager/internal/importer"
	"github.com/chankei613/mcp-server-manager/internal/mcp"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App はWailsアプリのメイン構造体。全Vueバインディングはここで定義する
type App struct {
	ctx     context.Context
	mu      sync.RWMutex
	clients map[uint]*clientEntry
	checker *health.Checker
}

type clientEntry struct {
	client    *mcp.Client
	transport mcp.Transport
	cancel    context.CancelFunc
}

func NewApp() *App {
	return &App{
		clients: make(map[uint]*clientEntry),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// データベース初期化
	dataDir := appDataDir()
	if err := db.Init(dataDir); err != nil {
		runtime.LogErrorf(ctx, "DB init failed: %s", err)
		return
	}

	// 前回セッションの connected/connecting をリセット（a.clients は空のため）
	db.DB.Model(&db.MCPServer{}).
		Where("status IN ?", []string{string(db.StatusConnected), string(db.StatusConnecting)}).
		Update("status", string(db.StatusDisconnected))

	// ヘルスチェッカー起動
	a.checker = health.NewChecker(func(serverID uint, eventType, message string) {
		a.emitEvent(serverID, eventType, message)
	})
	a.checker.Start(ctx)

	// アップデート確認（バックグラウンド）
	go func() {
		info, err := checkForUpdate()
		if err != nil {
			runtime.LogInfof(ctx, "update check skipped: %s", err)
			return
		}
		if info != nil {
			runtime.EventsEmit(ctx, "update:available", info)
		}
	}()

	runtime.LogInfo(ctx, "App started")
}

// GetAppVersion は現在のアプリバージョンを返す
func (a *App) GetAppVersion() string {
	return AppVersion
}

// ─── Server CRUD ──────────────────────────────────────────────────

// GetServers はDBのサーバー一覧をVueに返す
func (a *App) GetServers() ([]db.MCPServer, error) {
	var servers []db.MCPServer
	result := db.DB.Order("created_at asc").Find(&servers)
	return servers, result.Error
}

// AddServer は新しいMCPサーバー設定を追加する
func (a *App) AddServer(name, transport, command, argsJSON, url string) (*db.MCPServer, error) {
	server := db.MCPServer{
		Name:      name,
		Transport: db.TransportType(transport),
		Command:   command,
		Args:      argsJSON,
		URL:       url,
		Status:    db.StatusDisconnected,
	}
	result := db.DB.Create(&server)
	return &server, result.Error
}

// UpdateServer はサーバー設定を更新する（切断中のみ）
func (a *App) UpdateServer(id uint, name, command, argsJSON, url string) (*db.MCPServer, error) {
	var server db.MCPServer
	if err := db.DB.First(&server, id).Error; err != nil {
		return nil, err
	}
	if server.Status == db.StatusConnected {
		return nil, fmt.Errorf("disconnect the server before updating")
	}
	server.Name = name
	server.Command = command
	server.Args = argsJSON
	server.URL = url
	db.DB.Save(&server)
	return &server, nil
}

// DeleteServer はサーバーを削除する（切断してから削除）
func (a *App) DeleteServer(id uint) error {
	_ = a.DisconnectServer(id) // エラーは無視（既に切断済みの可能性）
	return db.DB.Delete(&db.MCPServer{}, id).Error
}

// ─── Connection lifecycle ─────────────────────────────────────────

// ConnectServer は指定サーバーへの接続を開始する
func (a *App) ConnectServer(id uint) error {
	var server db.MCPServer
	if err := db.DB.First(&server, id).Error; err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.clients[id]; exists {
		return nil // 既に接続中
	}

	ctx, cancel := context.WithCancel(a.ctx)

	onEvent := func(eventType, message string) {
		a.emitEvent(id, eventType, message)
	}

	switch server.Transport {
	case db.TransportStdio:
		args, err := importer.ParseArgs(server.Args)
		if err != nil {
			cancel()
			return fmt.Errorf("parse args: %w", err)
		}

		// Connect 時に元ファイルから env を取得（DB には保存しない）
		var envSlice []string
		if envMap := importer.FindServerEnv(server.Name); len(envMap) > 0 {
			for k, v := range envMap {
				envSlice = append(envSlice, k+"="+v)
			}
		}

		transport := mcp.NewStdioTransport(server.Command, args, envSlice, onEvent)
		if err := transport.Start(ctx); err != nil {
			cancel()
			a.updateStatus(id, db.StatusError, err.Error())
			return err
		}

		client := mcp.NewClient(transport, id)
		if _, err := client.Initialize(ctx); err != nil {
			transport.Stop()
			cancel()
			a.updateStatus(id, db.StatusError, err.Error())
			return fmt.Errorf("initialize: %w", err)
		}

		a.clients[id] = &clientEntry{client: client, transport: transport, cancel: cancel}
		a.checker.Register(id, &health.ServerEntry{Client: client, Transport: transport, Cancel: cancel})
		a.updateStatus(id, db.StatusConnected, "")

	case db.TransportHTTP:
		transport := mcp.NewHTTPTransport(server.URL, onEvent)
		if err := transport.Start(ctx); err != nil {
			cancel()
			a.updateStatus(id, db.StatusError, err.Error())
			return err
		}

		client := mcp.NewClient(transport, id)
		if _, err := client.Initialize(ctx); err != nil {
			transport.Stop()
			cancel()
			a.updateStatus(id, db.StatusError, err.Error())
			return fmt.Errorf("initialize: %w", err)
		}

		a.clients[id] = &clientEntry{client: client, transport: transport, cancel: cancel}
		a.checker.Register(id, &health.ServerEntry{Client: client, Transport: transport, Cancel: cancel})
		a.updateStatus(id, db.StatusConnected, "")

	default:
		cancel()
		return fmt.Errorf("unsupported transport: %s", server.Transport)
	}

	return nil
}

// DisconnectServer はサーバーとの接続を切断する
func (a *App) DisconnectServer(id uint) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	entry, exists := a.clients[id]
	if !exists {
		return nil
	}

	entry.cancel()
	entry.transport.Stop()
	delete(a.clients, id)
	a.checker.Unregister(id)
	a.updateStatus(id, db.StatusDisconnected, "")
	return nil
}

// ─── Tools ────────────────────────────────────────────────────────

// GetTools は接続中サーバーのツール一覧を取得する
func (a *App) GetTools(serverID uint) ([]mcp.Tool, error) {
	a.mu.RLock()
	entry, ok := a.clients[serverID]
	a.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("server %d is not connected", serverID)
	}
	return entry.client.ListTools(a.ctx)
}

// CallTool はツールを実行する
func (a *App) CallTool(serverID uint, toolName string, argumentsJSON string) (*mcp.ToolCallResult, error) {
	a.mu.RLock()
	entry, ok := a.clients[serverID]
	a.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("server %d is not connected", serverID)
	}

	var args json.RawMessage
	if argumentsJSON != "" {
		args = json.RawMessage(argumentsJSON)
	} else {
		args = json.RawMessage("{}")
	}

	return entry.client.CallTool(a.ctx, toolName, args)
}

// ─── Events ───────────────────────────────────────────────────────

// GetEvents は指定サーバーのイベントログを返す（最新100件）
func (a *App) GetEvents(serverID uint) ([]db.MCPEvent, error) {
	var events []db.MCPEvent
	result := db.DB.Where("server_id = ?", serverID).
		Order("created_at desc").
		Limit(100).
		Find(&events)
	return events, result.Error
}

// ─── Import ───────────────────────────────────────────────────────

// GetClaudeDesktopConfigPath はClaude Desktopの設定ファイルパスを返す
func (a *App) GetClaudeDesktopConfigPath() string {
	return importer.DefaultConfigPath()
}

// ImportClaudeDesktopConfig はClaude Desktopの設定をインポートする
func (a *App) ImportClaudeDesktopConfig(configPath string) (*importer.ImportResult, error) {
	result, err := importer.ImportFromClaudeDesktop(configPath)
	if err != nil {
		return nil, err
	}
	runtime.EventsEmit(a.ctx, "import:complete", result)
	return result, nil
}

// GetLocalClaudeConfigPath は ~/.claude/ 配下の MCP 設定ファイルパスを返す。
// claude_desktop_config.json が存在すればそちらを優先する。
func (a *App) GetLocalClaudeConfigPath() string {
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".claude.json"),                              // Claude Code のメイン設定（最優先）
		filepath.Join(home, ".claude", "claude_desktop_config.json"),
		filepath.Join(home, ".claude", "settings.local.json"),
		filepath.Join(home, ".claude", "settings.json"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			// mcpServers キーが含まれているか確認
			data, err := os.ReadFile(p)
			if err == nil && len(data) > 0 {
				if contains := string(data); len(contains) > 0 {
					// JSON内に mcpServers があるファイルを優先
					if strings.Contains(contains, `"mcpServers"`) {
						return p
					}
				}
			}
		}
	}
	// 見つからなければデフォルト
	return filepath.Join(home, ".claude", "settings.json")
}

// GetCursorConfigPath は ~/.cursor/mcp.json のパスを返す
func (a *App) GetCursorConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cursor", "mcp.json")
}

// GetWindsurfConfigPath は ~/.codeium/windsurf/mcp_config.json のパスを返す
func (a *App) GetWindsurfConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".codeium", "windsurf", "mcp_config.json")
}

// OpenFilePicker はOSのファイル選択ダイアログを開いて選択されたパスを返す
func (a *App) OpenFilePicker() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select MCP config file",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
		},
	})
	return path, err
}

// ─── Helpers ──────────────────────────────────────────────────────

func (a *App) emitEvent(serverID uint, eventType, message string) {
	// DBに記録
	db.DB.Create(&db.MCPEvent{
		ServerID:  serverID,
		EventType: eventType,
		Message:   message,
	})

	// Vueにリアルタイム送信
	runtime.EventsEmit(a.ctx, "server:event", map[string]interface{}{
		"serverID":  serverID,
		"eventType": eventType,
		"message":   message,
	})

	// サーバーステータスも更新通知
	var server db.MCPServer
	if db.DB.First(&server, serverID).Error == nil {
		runtime.EventsEmit(a.ctx, "server:status", server)
	}
}

func (a *App) updateStatus(serverID uint, status db.ServerStatus, errMsg string) {
	var server db.MCPServer
	if db.DB.First(&server, serverID).Error != nil {
		return
	}
	server.Status = status
	if errMsg != "" {
		server.LastError = errMsg
	}
	db.DB.Save(&server)

	runtime.EventsEmit(a.ctx, "server:status", server)
}

func appDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".mcp-server-manager")
}
