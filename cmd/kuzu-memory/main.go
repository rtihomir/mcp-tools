// File: cmd/kuzu-memory/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rtihomir/mcp-tools/internal/kuzu-memory/config"
	"github.com/rtihomir/mcp-tools/internal/kuzu-memory/db"
	"github.com/rtihomir/mcp-tools/internal/kuzu-memory/prompts"
	"github.com/rtihomir/mcp-tools/internal/kuzu-memory/tools"
)

func main() {
	// Parse configuration
	cfg, err := config.ParseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Initialize Kuzu database
	database, err := db.NewKuzuDB(cfg.DatabasePath, cfg.ReadOnly)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Setup signal handling for graceful shutdown (like JS version)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Fprintf(os.Stderr, "Received shutdown signal\n")
		os.Exit(0) // Exit like JS version
	}()

	// Create MCP server with both tools and prompts capabilities
	s := server.NewMCPServer(
		"kuzu-memory-server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(true),
	)

	// Register tools
	registerTools(s, database)

	// Register prompts
	registerPrompts(s, database)

	// Start stdio server
	fmt.Fprintf(os.Stderr, "Starting Kuzu MCP server (database: %s, read-only: %v)\n",
		cfg.DatabasePath, cfg.ReadOnly)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

// registerTools registers all MCP tools
func registerTools(s *server.MCPServer, database *db.KuzuDB) {
	// Register getSchema tool
	s.AddTool(mcp.NewTool("getSchema",
		mcp.WithDescription("Get the schema of the Kuzu database"),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleGetSchema(ctx, request, database)
	})

	// Register query tool
	s.AddTool(mcp.NewTool("query",
		mcp.WithDescription("Run a Cypher query on the Kuzu database"),
		mcp.WithString("cypher",
			mcp.Required(),
			mcp.Description("The Cypher query to run"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return tools.HandleQuery(ctx, request, database)
	})
}

// registerPrompts registers all MCP prompts
func registerPrompts(s *server.MCPServer, database *db.KuzuDB) {
	// Register generateKuzuCypher prompt
	s.AddPrompt(prompts.GetGenerateKuzuCypherPromptDefinition(),
		func(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return prompts.HandleGenerateKuzuCypher(request.Params.Arguments, database)
		})
}
