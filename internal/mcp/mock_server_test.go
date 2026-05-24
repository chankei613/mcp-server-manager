package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

// ─── pipe-based mock MCP server (stdio simulation) ───────────────────────────

type mockServer struct {
	serverIn  *os.File
	serverOut *os.File
	clientIn  *os.File
	clientOut *os.File
}

func newMockServer(t *testing.T) *mockServer {
	t.Helper()
	r0, w0, _ := os.Pipe() // client writes → server reads
	r1, w1, _ := os.Pipe() // server writes → client reads
	m := &mockServer{serverIn: r0, serverOut: w1, clientIn: r1, clientOut: w0}
	go m.serve()
	return m
}

func (m *mockServer) serve() {
	scanner := bufio.NewScanner(m.serverIn)
	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal([]byte(scanner.Text()), &req); err != nil {
			continue
		}
		var result interface{}
		switch req.Method {
		case "initialize":
			result = InitializeResult{
				ProtocolVersion: "2024-11-05",
				ServerInfo:      ServerInfo{Name: "mock-server", Version: "0.0.1"},
			}
		case "tools/list":
			result = ToolsListResult{Tools: []Tool{
				{Name: "echo", Description: "echoes input", InputSchema: json.RawMessage(`{"type":"object"}`)},
			}}
		case "tools/call":
			result = ToolCallResult{Content: []ContentBlock{{Type: "text", Text: "mock-result"}}}
		case "notifications/initialized":
			continue // no response for notifications
		default:
			resp := Response{JSONRPC: "2.0", ID: req.ID, Error: &RPCError{Code: -32601, Message: "unknown method"}}
			data, _ := json.Marshal(resp)
			fmt.Fprintf(m.serverOut, "%s\n", data)
			continue
		}
		raw, _ := json.Marshal(result)
		resp := Response{JSONRPC: "2.0", ID: req.ID, Result: raw}
		data, _ := json.Marshal(resp)
		fmt.Fprintf(m.serverOut, "%s\n", data)
	}
}

func (m *mockServer) stop() {
	m.clientOut.Close()
	m.serverOut.Close()
}

// ─── pipeTransport: Transport implementation over os.Pipe ────────────────────

type pipeTransport struct {
	r         *os.File
	w         *os.File
	mu        sync.Mutex
	pending   map[int64]chan *Response
	idCounter int64
	running   bool
}

func newPipeTransport(r, w *os.File) *pipeTransport {
	return &pipeTransport{r: r, w: w, pending: make(map[int64]chan *Response)}
}

func (t *pipeTransport) Start(_ context.Context) error {
	t.running = true
	go func() {
		scanner := bufio.NewScanner(t.r)
		for scanner.Scan() {
			var resp Response
			if err := json.Unmarshal([]byte(scanner.Text()), &resp); err != nil {
				continue
			}
			t.mu.Lock()
			ch, ok := t.pending[resp.ID]
			if ok {
				delete(t.pending, resp.ID)
			}
			t.mu.Unlock()
			if ok {
				ch <- &resp
			}
		}
	}()
	return nil
}

func (t *pipeTransport) Stop() error {
	t.running = false
	return t.w.Close()
}

func (t *pipeTransport) IsRunning() bool { return t.running }

func (t *pipeTransport) Call(_ context.Context, method string, params interface{}) (*Response, error) {
	t.mu.Lock()
	t.idCounter++
	id := t.idCounter
	var rawParams json.RawMessage
	if params != nil {
		raw, _ := json.Marshal(params)
		rawParams = raw
	}
	ch := make(chan *Response, 1)
	t.pending[id] = ch
	t.mu.Unlock()

	data, _ := json.Marshal(Request{JSONRPC: "2.0", ID: id, Method: method, Params: rawParams})
	fmt.Fprintf(t.w, "%s\n", data)

	select {
	case resp := <-ch:
		if resp.Error != nil {
			return nil, resp.Error
		}
		return resp, nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout waiting for %s", method)
	}
}

func (t *pipeTransport) Notify(method string, params interface{}) error {
	n := Notification{JSONRPC: "2.0", Method: method}
	if params != nil {
		raw, _ := json.Marshal(params)
		n.Params = raw
	}
	data, _ := json.Marshal(n)
	fmt.Fprintf(t.w, "%s\n", data)
	return nil
}

// ─── integration tests ───────────────────────────────────────────────────────

func TestClientInitialize(t *testing.T) {
	mock := newMockServer(t)
	defer mock.stop()

	tr := newPipeTransport(mock.clientIn, mock.clientOut)
	_ = tr.Start(context.Background())
	client := NewClient(tr, 1)

	result, err := client.Initialize(context.Background())
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	if result.ServerInfo.Name != "mock-server" {
		t.Errorf("got name %q, want mock-server", result.ServerInfo.Name)
	}
}

func TestClientListTools(t *testing.T) {
	mock := newMockServer(t)
	defer mock.stop()

	tr := newPipeTransport(mock.clientIn, mock.clientOut)
	_ = tr.Start(context.Background())
	client := NewClient(tr, 1)
	_, _ = client.Initialize(context.Background())

	tools, err := client.ListTools(context.Background())
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}
	if len(tools) != 1 || tools[0].Name != "echo" {
		t.Errorf("unexpected tools: %+v", tools)
	}
}

func TestClientCallTool(t *testing.T) {
	mock := newMockServer(t)
	defer mock.stop()

	tr := newPipeTransport(mock.clientIn, mock.clientOut)
	_ = tr.Start(context.Background())
	client := NewClient(tr, 1)
	_, _ = client.Initialize(context.Background())

	res, err := client.CallTool(context.Background(), "echo", json.RawMessage(`{"text":"hi"}`))
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if len(res.Content) == 0 || res.Content[0].Text != "mock-result" {
		t.Errorf("unexpected result: %+v", res)
	}
}

// ─── HTTP transport tests ─────────────────────────────────────────────────────

func TestHTTPTransportUnreachable(t *testing.T) {
	tr := NewHTTPTransport("http://127.0.0.1:19999", nil)
	err := tr.Start(context.Background())
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}

func TestHTTPTransportRoundtrip(t *testing.T) {
	// テスト用インラインHTTPサーバーを立てる
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(200)
			return
		}
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", 400)
			return
		}
		var result interface{}
		switch req.Method {
		case "initialize":
			result = InitializeResult{ProtocolVersion: "2024-11-05", ServerInfo: ServerInfo{Name: "http-mock"}}
		case "tools/list":
			result = ToolsListResult{Tools: []Tool{{Name: "ping", Description: "pong", InputSchema: json.RawMessage(`{}`)}}}
		default:
			result = map[string]string{}
		}
		raw, _ := json.Marshal(result)
		resp := Response{JSONRPC: "2.0", ID: req.ID, Result: raw}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln) //nolint:errcheck
	defer srv.Close()

	url := "http://" + ln.Addr().String() + "/mcp"
	tr := NewHTTPTransport(url, nil)
	if err := tr.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}

	client := NewClient(tr, 99)
	info, err := client.Initialize(context.Background())
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if info.ServerInfo.Name != "http-mock" {
		t.Errorf("got %q, want http-mock", info.ServerInfo.Name)
	}

	tools, err := client.ListTools(context.Background())
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if len(tools) != 1 || tools[0].Name != "ping" {
		t.Errorf("unexpected tools: %+v", tools)
	}
}
