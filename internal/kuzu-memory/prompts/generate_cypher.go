package prompts

import (
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rtihomir/mcp-tools/internal/kuzu-memory/db"
)

// HandleGenerateKuzuCypher handles the generateKuzuCypher prompt
func HandleGenerateKuzuCypher(arguments map[string]string, db *db.KuzuDB) (*mcp.GetPromptResult, error) {
	// Extract question parameter
	question, exists := arguments["question"]
	if !exists {
		return nil, fmt.Errorf("missing required parameter: question")
	}

	if question == "" {
		return nil, fmt.Errorf("question parameter cannot be empty")
	}

	// Get current schema
	schema, err := db.GetSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	// Generate the prompt content
	promptContent, err := generatePromptContent(question, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prompt: %w", err)
	}

	return &mcp.GetPromptResult{
		Messages: []mcp.PromptMessage{
			{
				Role: "user",
				Content: mcp.TextContent{
					Type: "text",
					Text: promptContent,
				},
			},
		},
	}, nil
}

// generatePromptContent creates the exact same prompt as the JS version
func generatePromptContent(question string, schema *db.Schema) (string, error) {
	// Convert schema to JSON (exactly like JS version)
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}

	// Build the exact prompt from JS version
	prompt := fmt.Sprintf(`Task:Generate Kuzu Cypher statement to query a graph database.
Instructions:
Generate the Kuzu dialect of Cypher with the following rules in mind:
1. It is recommended to always specifying node and relationship labels explicitly in the CREATE and MERGE clause. If not specified, Kuzu will try to infer the label by looking at the schema.
2. FINISH is recently introduced in GQL and adopted by Neo4j but not yet supported in Kuzu. You can use RETURN COUNT(*) instead which will only return one record.
3. FOREACH is not supported. You can use UNWIND instead.
4. Kuzu can scan files not only in the format of CSV, so the LOAD CSV FROM clause is renamed to LOAD FROM.
5. Relationship cannot be omitted. For example --, -- >  and < -- are not supported. You need to use  - [] - ,  - [] ->  and  < -[] - instead.
6. Neo4j adopts trail semantic (no repeated edge) for pattern within a MATCH clause. While Kuzu adopts walk semantic (allow repeated edge) for pattern within a MATCH clause. You can use is_trail or is_acyclic function to check if a path is a trail or acyclic.
7. Since Kuzu adopts trail semantic by default, so a variable length relationship needs to have a upper bound to guarantee the query will terminate. If upper bound is not specified, Kuzu will assign a default value of 30.
8. To run algorithms like (all) shortest path, simply add SHORTEST or ALL SHORTEST between the kleene star and lower bound. For example,  MATCH(n) - [r * SHORTEST 1..10] -> (m). It is recommended to use SHORTEST if paths are not needed in the use case.
9. REMOVE is not supported. Use SET n.prop = NULL instead.
10. Properties must be updated in the form of n.prop = expression. Update all properties with map of  += operator is not supported. Try to update properties one by one.
11. USE graph is not supported. For Kuzu, each graph is a database.
12. Using WHERE inside node or relationship pattern is not supported, e.g. MATCH(n: Person WHERE a.name = 'Andy') RETURN n. You need to write it as MATCH(n: Person) WHERE n.name = 'Andy' RETURN n.
13. Filter on node or relationship labels is not supported, e.g. MATCH (n) WHERE n:Person RETURN n. You need to write it as MATCH(n: Person) RETURN n, or MATCH(n) WHERE label(n) = 'Person' RETURN n.
14. Any SHOW XXX clauses become a function call in Kuzu. For example, SHOW FUNCTIONS in Neo4j is equivalent to CALL show_functions() RETURN * in Kuzu.
15. Kuzu supports EXISTS and COUNT subquery.
16. CALL <subquery> is not supported.

Use only the provided node types, relationship types and properties in the schema.
Do not use any other node types, relationship types or properties that are not provided explicitly in the schema.
Schema:
%s
Note: Do not include any explanations or apologies in your responses.
Do not respond to any questions that might ask anything else than for you to construct a Cypher statement.
Do not include any text except the generated Cypher statement.

The question is:
%s
`, string(schemaJSON), question)

	return prompt, nil
}

// GetGenerateKuzuCypherPromptDefinition returns the prompt definition for MCP registration
func GetGenerateKuzuCypherPromptDefinition() mcp.Prompt {
	return mcp.Prompt{
		Name:        "generateKuzuCypher",
		Description: "Generate a Cypher query for Kuzu",
		Arguments: []mcp.PromptArgument{
			{
				Name:        "question",
				Description: "The question in natural language to generate the Cypher query for",
				Required:    true,
			},
		},
	}
}
