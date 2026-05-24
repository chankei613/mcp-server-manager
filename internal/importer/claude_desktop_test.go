package importer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseArgs(t *testing.T) {
	args, err := ParseArgs(`["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]`)
	require.NoError(t, err)
	assert.Equal(t, []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"}, args)
}

func TestParseArgsEmpty(t *testing.T) {
	args, err := ParseArgs("[]")
	require.NoError(t, err)
	assert.Empty(t, args)
}

func TestParseArgsFallback(t *testing.T) {
	// invalid JSON falls back to space-split
	args, err := ParseArgs("npx -y server")
	require.NoError(t, err)
	assert.Equal(t, []string{"npx", "-y", "server"}, args)
}

func TestImportFromClaudeDesktop(t *testing.T) {
	config := `{
		"mcpServers": {
			"filesystem": {
				"command": "npx",
				"args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
			},
			"http-server": {
				"url": "http://localhost:8080"
			}
		}
	}`

	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "claude_desktop_config.json")
	require.NoError(t, os.WriteFile(configPath, []byte(config), 0644))

	// point DB to temp dir (will fail DB operations without real DB, but tests parsing)
	// We just test that it reads the file correctly via parseConfig
	cfg, err := parseConfig(configPath)
	require.NoError(t, err)
	assert.Len(t, cfg.MCPServers, 2)
	assert.Equal(t, "npx", cfg.MCPServers["filesystem"].Command)
	assert.Equal(t, "http://localhost:8080", cfg.MCPServers["http-server"].URL)
}
