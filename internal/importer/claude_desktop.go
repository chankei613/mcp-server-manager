package importer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/chankei613/mcp-server-manager/internal/db"
)

// ClaudeDesktopConfig はclaude_desktop_config.jsonの構造
type ClaudeDesktopConfig struct {
	MCPServers map[string]ClaudeDesktopServer `json:"mcpServers"`
}

type ClaudeDesktopServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	URL     string            `json:"url,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

func parseConfig(configPath string) (*ClaudeDesktopConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg ClaudeDesktopConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// DefaultConfigPath はOSに応じたClaude Desktopの設定ファイルパスを返す
func DefaultConfigPath() string {
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	case "windows":
		appData := os.Getenv("APPDATA")
		return filepath.Join(appData, "Claude", "claude_desktop_config.json")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", "claude", "claude_desktop_config.json")
	}
}

// ImportResult はインポートの結果サマリー
type ImportResult struct {
	Imported []string `json:"imported"`
	Skipped  []string `json:"skipped"` // 既に存在する
	Errors   []string `json:"errors"`
}

// ImportFromClaudeDesktop は claude_desktop_config.json を読み込んでDBに保存する
func ImportFromClaudeDesktop(configPath string) (*ImportResult, error) {
	cfg, err := parseConfig(configPath)
	if err != nil {
		return nil, err
	}

	result := &ImportResult{}

	for name, server := range cfg.MCPServers {
		// 既存チェック
		var existing db.MCPServer
		if tx := db.DB.Where("name = ?", name).First(&existing); tx.Error == nil {
			result.Skipped = append(result.Skipped, name)
			continue
		}

		transport := db.TransportStdio
		if server.URL != "" {
			transport = db.TransportHTTP
		}

		argsJSON, err := json.Marshal(server.Args)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", name, err))
			continue
		}

		newServer := db.MCPServer{
			Name:      name,
			Transport: transport,
			Command:   server.Command,
			Args:      string(argsJSON),
			URL:       server.URL,
			Status:    db.StatusDisconnected,
		}

		if tx := db.DB.Create(&newServer); tx.Error != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", name, tx.Error))
			continue
		}

		result.Imported = append(result.Imported, name)
	}

	return result, nil
}

// FindServerEnv は全既知設定ファイルをサーバー名で検索し、envマップを返す（大文字小文字無視）
// セキュリティ方針: env は DB に保存せず Connect 時にのみ取得する
func FindServerEnv(serverName string) map[string]string {
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".claude.json"),
		filepath.Join(home, ".claude", "claude_desktop_config.json"),
		filepath.Join(home, ".claude", "settings.local.json"),
		filepath.Join(home, ".claude", "settings.json"),
		filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"),
		filepath.Join(home, ".cursor", "mcp.json"),
		filepath.Join(home, ".codeium", "windsurf", "mcp_config.json"),
	}
	lowerName := strings.ToLower(serverName)
	for _, path := range candidates {
		cfg, err := parseConfig(path)
		if err != nil {
			continue
		}
		// 完全一致を優先し、次に大文字小文字無視で検索
		if server, ok := cfg.MCPServers[serverName]; ok && len(server.Env) > 0 {
			return server.Env
		}
		for key, server := range cfg.MCPServers {
			if strings.ToLower(key) == lowerName && len(server.Env) > 0 {
				return server.Env
			}
		}
	}
	return nil
}

// ParseArgs はJSON配列文字列を[]stringに変換する
func ParseArgs(argsJSON string) ([]string, error) {
	if argsJSON == "" || argsJSON == "null" {
		return nil, nil
	}
	var args []string
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		// スペース区切りの文字列としてフォールバック
		return strings.Fields(argsJSON), nil
	}
	return args, nil
}
