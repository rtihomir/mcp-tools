#!/bin/bash

echo "=== Testing DuckDB MCP Server ==="

# Function to send JSON-RPC request
send_request() {
    echo "$1" | ./build/duckdb 2>/dev/null | head -1 | jq .
}

# Initialize server
echo "1. Initializing server..."
INIT='{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {"tools": {}}, "clientInfo": {"name": "test", "version": "1.0"}}}'
send_request "$INIT"

echo -e "\n2. Testing configure tool with home directory..."
CONFIG_HOME='{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "configure", "arguments": {"home_dir": "/tmp/duckdb-test"}}}'
send_request "$INIT"$'\n'"$CONFIG_HOME"

echo -e "\n3. Testing list_files tool..."
LIST_FILES='{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "list_files", "arguments": {}}}'
send_request "$INIT"$'\n'"$CONFIG_HOME"$'\n'"$LIST_FILES"

echo -e "\n4. Testing configure with CSV file..."
CONFIG_CSV='{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "configure", "arguments": {"db_path": "/tmp/duckdb-test/people.csv"}}}'
send_request "$INIT"$'\n'"$CONFIG_CSV"

echo -e "\n5. Testing query on CSV..."
QUERY_CSV='{"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "query", "arguments": {"sql": "SELECT * FROM '"'"'/tmp/duckdb-test/people.csv'"'"'"}}}'
send_request "$INIT"$'\n'"$CONFIG_CSV"$'\n'"$QUERY_CSV"

echo -e "\n6. Testing memory database..."
CONFIG_MEM='{"jsonrpc": "2.0", "id": 6, "method": "tools/call", "params": {"name": "configure", "arguments": {"db_path": ":memory:"}}}'
send_request "$INIT"$'\n'"$CONFIG_MEM"

echo -e "\n7. Testing query in memory database..."
QUERY_MEM='{"jsonrpc": "2.0", "id": 7, "method": "tools/call", "params": {"name": "query", "arguments": {"sql": "SELECT 1 + 1 AS result"}}}'
send_request "$INIT"$'\n'"$CONFIG_MEM"$'\n'"$QUERY_MEM"

echo -e "\n=== Test Complete ==="