package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer("Current DateTime", "1.0.0")

	tool := mcp.NewTool("get_current_date_time",
		mcp.WithDescription("Returns current Date and Time in ISO 8601 format with time zone offset"),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		isoTime := time.Now().Format(time.RFC3339)
		return mcp.NewToolResultText(isoTime), nil
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
