package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rtihomir/mcp-tools/internal/kuzu-memory/db"
)

// HandleQuery implements the query tool exactly like the JS version
func HandleQuery(ctx context.Context, request mcp.CallToolRequest, db *db.KuzuDB) (*mcp.CallToolResult, error) {
	// Type assert Arguments to map
	arguments, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid arguments format")
	}

	// Extract cypher parameter
	cypherParam, exists := arguments["cypher"]
	if !exists {
		return nil, fmt.Errorf("missing required parameter: cypher")
	}

	cypher, ok := cypherParam.(string)
	if !ok {
		return nil, fmt.Errorf("cypher parameter must be a string")
	}

	// Execute query and get all results (mimics JS queryResult.getAll())
	rows, err := db.ExecuteQueryAndGetAll(cypher)
	if err != nil {
		// In JS version, errors are thrown directly, so we do the same
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	// Convert to JSON with 2-space indentation (exactly like JS)
	jsonResult, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query result: %w", err)
	}

	return mcp.NewToolResultText(string(jsonResult)), nil
}
