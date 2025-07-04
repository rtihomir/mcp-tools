package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rtihomir/mcp-tools/internal/duckdb/prompt"
	"github.com/rtihomir/mcp-tools/internal/duckdb/state"
	"github.com/rtihomir/mcp-tools/internal/duckdb/tools"
)

func main() {
	// Create session for managing state
	session := state.NewSession()
	defer session.Close()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Fprintf(os.Stderr, "Received shutdown signal\n")
		session.Close()
		os.Exit(0)
	}()

	// Create MCP server with tools and prompts capabilities
	s := server.NewMCPServer(
		"duckdb-server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(true),
	)

	// Register tools
	registerTools(s, session)

	// Register prompts
	registerPrompts(s)

	// Start stdio server
	fmt.Fprintf(os.Stderr, "Starting DuckDB MCP server\n")

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

// registerTools registers all MCP tools
func registerTools(s *server.MCPServer, session *state.Session) {
	// Register configure tool
	s.AddTool(tools.GetConfigureToolDefinition(),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return tools.HandleConfigure(ctx, request, session)
		})

	// Register query tool
	s.AddTool(tools.GetQueryToolDefinition(),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return tools.HandleQuery(ctx, request, session)
		})

	// Register list_files tool
	s.AddTool(tools.GetListFilesToolDefinition(),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return tools.HandleListFiles(ctx, request, session)
		})
}

// registerPrompts registers all MCP prompts
func registerPrompts(s *server.MCPServer) {
	// Register DuckDB initial prompt
	s.AddPrompt(prompt.GetDuckDBInitialPromptDefinition(),
		func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return prompt.HandleDuckDBInitialPrompt(request.Params.Arguments)
		})
}
