// File: internal/kuzu-memory/tools/schema.go
package tools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rtihomir/mcp-tools/internal/kuzu-memory/db"
)

// HandleGetSchema implements the getSchema tool exactly like the JS version
func HandleGetSchema(ctx context.Context, request mcp.CallToolRequest, db *db.KuzuDB) (*mcp.CallToolResult, error) {
	// Get schema from database
	schema, err := db.GetSchema()
	if err != nil {
		return nil, err
	}

	// Convert to JSON with 2-space indentation (exactly like JS)
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(schemaJSON)), nil
}
