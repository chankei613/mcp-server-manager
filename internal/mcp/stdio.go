package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// StdioTransport はstdioベースのMCPサーバー（子プロセス）を管理する
type StdioTransport struct {
	command string
	args    []string
	env     []string // "KEY=VALUE" 形式。元の環境変数にマージして渡す

	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner

	mu      sync.Mutex
	pending map[int64]chan *Response

	idCounter atomic.Int64
	running   atomic.Bool

	onEvent func(eventType, message string)
}

func NewStdioTransport(command string, args []string, env []string, onEvent func(string, string)) *StdioTransport {
	return &StdioTransport{
		command: command,
		args:    args,
		env:     env,
		pending: make(map[int64]chan *Response),
		onEvent: onEvent,
	}
}

// Start は子プロセスを起動し、レスポンス読み取りループを開始する
func (t *StdioTransport) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running.Load() {
		return nil
	}

	t.cmd = exec.CommandContext(ctx, t.command, t.args...)
	// 現在の環境変数に追加の env をマージ（PATH等を引き継ぎつつ API キーを渡す）
	if len(t.env) > 0 {
		t.cmd.Env = append(os.Environ(), t.env...)
	}

	var err error
	t.stdin, err = t.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}

	stdout, err := t.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	t.stdout = bufio.NewScanner(stdout)

	// stderrはイベントとしてキャプチャ
	stderr, _ := t.cmd.StderrPipe()

	if err := t.cmd.Start(); err != nil {
		return fmt.Errorf("start process: %w", err)
	}

	t.running.Store(true)
	t.emit("connected", fmt.Sprintf("started process: %s %s", t.command, strings.Join(t.args, " ")))

	go t.readLoop()
	go t.readStderr(stderr)
	go t.waitProcess()

	return nil
}

// Stop は子プロセスを終了する
func (t *StdioTransport) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running.Load() {
		return nil
	}

	t.running.Store(false)
	if t.stdin != nil {
		t.stdin.Close()
	}
	if t.cmd != nil && t.cmd.Process != nil {
		t.cmd.Process.Kill()
	}

	// 待機中のリクエストを全てエラーで解決
	for id, ch := range t.pending {
		ch <- &Response{Error: &RPCError{Code: -1, Message: "server stopped"}}
		delete(t.pending, id)
	}

	t.emit("disconnected", "server stopped")
	return nil
}

// Call はJSON-RPC 2.0リクエストを送り、レスポンスを待つ
func (t *StdioTransport) Call(ctx context.Context, method string, params interface{}) (*Response, error) {
	if !t.running.Load() {
		return nil, fmt.Errorf("server not running")
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

	req := Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  rawParams,
	}

	ch := make(chan *Response, 1)
	t.mu.Lock()
	t.pending[id] = ch
	t.mu.Unlock()

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	t.mu.Lock()
	_, err = fmt.Fprintf(t.stdin, "%s\n", data)
	t.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("write to stdin: %w", err)
	}

	// タイムアウト付きで待機
	timeout := 30 * time.Second
	select {
	case resp := <-ch:
		if resp.Error != nil {
			return nil, resp.Error
		}
		return resp, nil
	case <-time.After(timeout):
		t.mu.Lock()
		delete(t.pending, id)
		t.mu.Unlock()
		return nil, fmt.Errorf("request timeout after %s", timeout)
	case <-ctx.Done():
		t.mu.Lock()
		delete(t.pending, id)
		t.mu.Unlock()
		return nil, ctx.Err()
	}
}

// Notify はレスポンスを待たない通知を送る（initialized等）
func (t *StdioTransport) Notify(method string, params interface{}) error {
	n := Notification{JSONRPC: "2.0", Method: method}
	if params != nil {
		raw, err := json.Marshal(params)
		if err != nil {
			return err
		}
		n.Params = raw
	}
	data, err := json.Marshal(n)
	if err != nil {
		return err
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	_, err = fmt.Fprintf(t.stdin, "%s\n", data)
	return err
}

func (t *StdioTransport) IsRunning() bool {
	return t.running.Load()
}

func (t *StdioTransport) readLoop() {
	for t.stdout.Scan() {
		line := strings.TrimSpace(t.stdout.Text())
		if line == "" {
			continue
		}

		var resp Response
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
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
}

func (t *StdioTransport) readStderr(r io.ReadCloser) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			t.emit("stderr", line)
		}
	}
}

func (t *StdioTransport) waitProcess() {
	if err := t.cmd.Wait(); err != nil && t.running.Load() {
		t.running.Store(false)
		t.emit("error", fmt.Sprintf("process exited: %s", err))
	}
}

func (t *StdioTransport) emit(eventType, message string) {
	if t.onEvent != nil {
		t.onEvent(eventType, message)
	}
}
