package mcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestMarshal(t *testing.T) {
	req := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}
	data, err := json.Marshal(req)
	require.NoError(t, err)

	var back Request
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, "2.0", back.JSONRPC)
	assert.Equal(t, int64(1), back.ID)
	assert.Equal(t, "tools/list", back.Method)
}

func TestResponseError(t *testing.T) {
	raw := `{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"method not found"}}`
	var resp Response
	require.NoError(t, json.Unmarshal([]byte(raw), &resp))
	assert.NotNil(t, resp.Error)
	assert.Equal(t, -32601, resp.Error.Code)
	assert.Equal(t, "method not found", resp.Error.Message)
}

func TestRPCErrorImplementsError(t *testing.T) {
	rpcErr := &RPCError{Code: -32600, Message: "invalid request"}
	assert.Contains(t, rpcErr.Error(), "invalid request")
	assert.Contains(t, rpcErr.Error(), "-32600")
}

func TestNotificationMarshal(t *testing.T) {
	n := Notification{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	data, err := json.Marshal(n)
	require.NoError(t, err)
	// notifications have no id field
	assert.NotContains(t, string(data), `"id"`)
	assert.Contains(t, string(data), `"notifications/initialized"`)
}
