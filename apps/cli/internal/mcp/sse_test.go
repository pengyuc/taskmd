package mcp

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestSSEServer(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	origWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origWd)

	// Create a dummy task file
	taskFile := "001-test.md"
	content := `---
id: "001"
title: "Test Task"
status: pending
---
# Test Task`
	if err := os.WriteFile(taskFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write task file: %v", err)
	}

	// Create MCP server
	server := NewServer("test")

	// Create SSE handler
	handler := gomcp.NewSSEHandler(func(*http.Request) *gomcp.Server { return server }, nil)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Create client with SSE transport
	ctx := context.Background()
	transport := &gomcp.SSEClientTransport{Endpoint: ts.URL}
	client := gomcp.NewClient(&gomcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	
	cs, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("failed to connect to SSE server: %v", err)
	}
	defer cs.Close()

	// Call a tool (e.g., list)
	res, err := cs.CallTool(ctx, &gomcp.CallToolParams{
		Name:      "list",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("failed to call list: %v", err)
	}

	// Verify result
	if len(res.Content) == 0 {
		t.Fatal("expected content in result, got none")
	}
	
	text, ok := res.Content[0].(*gomcp.TextContent)
	if !ok {
		t.Fatalf("expected text content, got %T", res.Content[0])
	}

	if text.Text == "" {
		t.Fatal("expected non-empty text in result")
	}

	fmt.Printf("SSE Test Result: %s\n", text.Text)
}
