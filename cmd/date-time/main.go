package main

import (
	"context"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Create a new MCP server
	s := server.NewMCPServer(
		"date-time-server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Register the current_date_time tool
	s.AddTool(mcp.NewTool("current_date_time",
		mcp.WithDescription("Returns the current date and time in ISO 8601 format"),
	), handleCurrentDateTime)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		panic(err)
	}
}

func handleCurrentDateTime(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get current date and time in ISO 8601 format
	currentTime := time.Now().Format(time.RFC3339)

	// Return the result
	return mcp.NewToolResultText(currentTime), nil
}
