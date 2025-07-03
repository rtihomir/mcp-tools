package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rtihomir/mcp-tools/internal/duckdb/state"
)

// QueryRequest represents the parameters for the query tool
type QueryRequest struct {
	SQL string `json:"sql"`
}

// QueryResponse represents the response from the query tool
type QueryResponse struct {
	Success bool   `json:"success"`
	Results string `json:"results,omitempty"`
	Error   string `json:"error,omitempty"`
	DBPath  string `json:"db_path"`
}

// HandleQuery handles the query tool request
func HandleQuery(ctx context.Context, request mcp.CallToolRequest, session *state.Session) (*mcp.CallToolResult, error) {
	// Check if session has a database connection
	if !session.HasConnection() {
		return mcp.NewToolResultError("No database connection. Use the 'configure' tool first to connect to a database."), nil
	}

	// Parse the request arguments
	var req QueryRequest
	if request.Params.Arguments == nil {
		return mcp.NewToolResultError("Missing required parameter: 'sql'"), nil
	}

	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		if err := mapToStruct(args, &req); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid arguments: %v", err)), nil
		}
	} else {
		return mcp.NewToolResultError("Invalid argument format"), nil
	}

	if req.SQL == "" {
		return mcp.NewToolResultError("Parameter 'sql' cannot be empty"), nil
	}

	// Get the database client
	client := session.GetClient()
	if client == nil {
		return mcp.NewToolResultError("Database connection is not available"), nil
	}

	// Execute the query
	results, err := client.Query(req.SQL)
	
	response := QueryResponse{
		DBPath: session.GetDBPath(),
	}

	if err != nil {
		response.Success = false
		response.Error = err.Error()
	} else {
		response.Success = true
		response.Results = results
	}

	// Convert response to JSON
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(responseBytes)), nil
}

// GetQueryToolDefinition returns the MCP tool definition for query
func GetQueryToolDefinition() mcp.Tool {
	return mcp.NewTool("query",
		mcp.WithDescription("Execute a SQL query on the configured DuckDB database"),
		mcp.WithString("sql",
			mcp.Required(),
			mcp.Description("SQL query to execute (DuckDB dialect)"),
		),
	)
}