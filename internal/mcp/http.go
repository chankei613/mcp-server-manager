package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// HTTPTransport はHTTP経由でJSON-RPC 2.0リクエストを送るMCPトランスポート
type HTTPTransport struct {
	url     string
	client  *http.Client
	running atomic.Bool
	onEvent func(eventType, message string)

	idCounter atomic.Int64
}

func NewHTTPTransport(url string, onEvent func(string, string)) *HTTPTransport {
	return &HTTPTransport{
		url:     url,
		onEvent: onEvent,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *HTTPTransport) Start(ctx context.Context) error {
	// 疎通確認: initialize を呼ぶ前に一度 ping 代わりに接続確認
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.url, nil)
	if err != nil {
		return fmt.Errorf("build ping request: %w", err)
	}
	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("connect to %s: %w", t.url, err)
	}
	resp.Body.Close()

	t.running.Store(true)
	t.emit("connected", fmt.Sprintf("connected to %s", t.url))
	return nil
}

func (t *HTTPTransport) Stop() error {
	t.running.Store(false)
	t.emit("disconnected", "HTTP transport stopped")
	return nil
}

func (t *HTTPTransport) Call(ctx context.Context, method string, params interface{}) (*Response, error) {
	if !t.running.Load() {
		return nil, fmt.Errorf("transport not running")
	}

	id := t.idCounter.Add(1)

	var rawParams json.RawMessage
	if params != nil {
		var err error
		rawParams, err = json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("marshal params: %w", err)
		}
	}

	reqBody := Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  rawParams,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, t.url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("http %d from %s", httpResp.StatusCode, t.url)
	}

	var rpcResp Response
	if err := json.NewDecoder(httpResp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, rpcResp.Error
	}

	return &rpcResp, nil
}

func (t *HTTPTransport) Notify(method string, params interface{}) error {
	if !t.running.Load() {
		return nil
	}

	n := Notification{JSONRPC: "2.0", Method: method}
	if params != nil {
		raw, err := json.Marshal(params)
		if err != nil {
			return err
		}
		n.Params = raw
	}

	body, err := json.Marshal(n)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (t *HTTPTransport) IsRunning() bool {
	return t.running.Load()
}

func (t *HTTPTransport) emit(eventType, message string) {
	if t.onEvent != nil {
		t.onEvent(eventType, message)
	}
}
