package cli

import (
	"context"
	"fmt"
	"net/http"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	taskmcp "github.com/driangle/taskmd/apps/cli/internal/mcp"
)

var (
	mcpTransport string
	mcpPort      int
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server",
	Long: `Start a Model Context Protocol (MCP) server that communicates over stdin/stdout or HTTP (SSE).

This allows LLM-based tools (Cursor, Windsurf, Copilot agents, etc.) to interact
with your taskmd project using the standard MCP protocol.

Example configuration for Claude Code (.mcp.json):
  {
    "mcpServers": {
      "taskmd": {
        "command": "taskmd",
        "args": ["mcp"]
      }
    }
  }

To start as an HTTP server:
  taskmd mcp --transport sse --port 8080`,
	Args: cobra.NoArgs,
	RunE: runMcp,
}

func init() {
	rootCmd.AddCommand(mcpCmd)

	mcpCmd.Flags().StringVar(&mcpTransport, "transport", "stdio", "Transport protocol to use (stdio, sse)")
	mcpCmd.Flags().IntVar(&mcpPort, "port", 8080, "Port to listen on for SSE transport")
}

func runMcp(_ *cobra.Command, _ []string) error {
	server := taskmcp.NewServer(Version)

	switch mcpTransport {
	case "stdio":
		return server.Run(context.Background(), &gomcp.StdioTransport{})
	case "sse":
		handler := gomcp.NewSSEHandler(func(*http.Request) *gomcp.Server { return server }, nil)
		addr := fmt.Sprintf(":%d", mcpPort)
		fmt.Printf("Starting MCP SSE server on %s\n", addr)
		return http.ListenAndServe(addr, handler)
	default:
		return fmt.Errorf("unsupported transport: %s", mcpTransport)
	}
}
