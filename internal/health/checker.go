package health

import (
	"context"
	"sync"
	"time"

	"github.com/chankei613/mcp-server-manager/internal/db"
	"github.com/chankei613/mcp-server-manager/internal/mcp"
)

const (
	checkInterval       = 30 * time.Second
	circuitBreakerLimit = 3 // 3回連続失敗でdegraded
)

// ServerEntry はヘルスチェック対象のサーバー情報
type ServerEntry struct {
	Client    *mcp.Client
	Transport *mcp.StdioTransport
	Cancel    context.CancelFunc
}

// Checker はすべてのMCPサーバーのヘルスチェックを管理する
type Checker struct {
	mu      sync.RWMutex
	servers map[uint]*ServerEntry
	onEvent func(serverID uint, eventType, message string)
}

func NewChecker(onEvent func(uint, string, string)) *Checker {
	return &Checker{
		servers: make(map[uint]*ServerEntry),
		onEvent: onEvent,
	}
}

// Register はサーバーをヘルスチェック対象に追加する
func (c *Checker) Register(serverID uint, entry *ServerEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.servers[serverID] = entry
}

// Unregister はサーバーをヘルスチェックから除外する
func (c *Checker) Unregister(serverID uint) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if entry, ok := c.servers[serverID]; ok {
		entry.Cancel()
		delete(c.servers, serverID)
	}
}

// Start はヘルスチェックのメインループを起動する（goroutine）
func (c *Checker) Start(ctx context.Context) {
	ticker := time.NewTicker(checkInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.checkAll(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (c *Checker) checkAll(ctx context.Context) {
	c.mu.RLock()
	servers := make(map[uint]*ServerEntry, len(c.servers))
	for id, entry := range c.servers {
		servers[id] = entry
	}
	c.mu.RUnlock()

	for serverID, entry := range servers {
		go c.check(ctx, serverID, entry)
	}
}

func (c *Checker) check(ctx context.Context, serverID uint, entry *ServerEntry) {
	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// tools/listを送って疎通確認（軽量なpingの代わり）
	_, err := entry.Client.ListTools(checkCtx)

	var server db.MCPServer
	if result := db.DB.First(&server, serverID); result.Error != nil {
		return
	}

	now := time.Now()

	if err != nil {
		server.ConsecutiveFailures++
		server.LastError = err.Error()
		server.LastErrorAt = &now
		c.onEvent(serverID, "error", err.Error())

		if server.ConsecutiveFailures >= circuitBreakerLimit {
			server.Status = db.StatusDegraded
			c.onEvent(serverID, "degraded", "circuit breaker triggered after 3 consecutive failures")
		}
	} else {
		if server.Status != db.StatusConnected {
			c.onEvent(serverID, "connected", "health check passed")
		}
		server.Status = db.StatusConnected
		server.ConsecutiveFailures = 0
		server.LastConnectedAt = &now
	}

	db.DB.Save(&server)
}
