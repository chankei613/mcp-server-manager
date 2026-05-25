package mcp

import (
	"context"
	"strings"
	"testing"
)

// TestStdioTransportProcessExitCancelsRequest verifies that when a subprocess exits
// unexpectedly, pending requests fail immediately instead of waiting for the 30s timeout.
// This covers the bug: 'initialize: request timeout after 30s' when GITHUB_PERSONAL_ACCESS_TOKEN
// is missing and the server process exits right away.
func TestStdioTransportProcessExitCancelsRequest(t *testing.T) {
	// 'true' exits immediately with code 0, simulating a server that exits without responding
	tr := NewStdioTransport("true", nil, nil, nil)
	if err := tr.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// This call should return quickly (process exited), NOT hang for 30s
	done := make(chan error, 1)
	go func() {
		_, err := tr.Call(context.Background(), "initialize", nil)
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error when process exits, got nil")
		}
		if !strings.Contains(err.Error(), "process exited") && !strings.Contains(err.Error(), "server not running") {
			t.Errorf("expected 'process exited' error, got: %v", err)
		}
	case <-context.Background().Done():
	}
}

// TestStdioTransportProcessExitWithStderr verifies that stderr output from a failing
// process is captured as an event.
func TestStdioTransportProcessExitWithStderr(t *testing.T) {
	var events []string
	onEvent := func(eventType, msg string) {
		events = append(events, eventType+":"+msg)
	}

	// 'sh -c "echo error>&2; exit 1"' writes to stderr and exits with error
	tr := NewStdioTransport("sh", []string{"-c", "echo 'missing token' >&2; exit 1"}, nil, onEvent)
	if err := tr.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Wait for process to exit
	_, _ = tr.Call(context.Background(), "initialize", nil)

	// Should have captured stderr and error event
	found := false
	for _, e := range events {
		if strings.Contains(e, "missing token") || strings.Contains(e, "error") || strings.Contains(e, "process exited") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error/stderr event, got events: %v", events)
	}
}
