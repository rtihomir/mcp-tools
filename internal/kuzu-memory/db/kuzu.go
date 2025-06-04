package db

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/kuzudb/go-kuzu"
)

const (
	TABLE_TYPE_NODE = "NODE"
	TABLE_TYPE_REL  = "REL"
)

// Property represents a table property
type Property struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	IsPrimaryKey bool   `json:"isPrimaryKey,omitempty"` // omitempty for rel tables
}

// Connectivity represents relationship connectivity
type Connectivity struct {
	Src string `json:"src"`
	Dst string `json:"dst"`
}

// NodeTable represents a node table in schema
type NodeTable struct {
	Name       string     `json:"name"`
	Comment    string     `json:"comment"`
	Properties []Property `json:"properties"`
}

// RelTable represents a relationship table in schema
type RelTable struct {
	Name         string         `json:"name"`
	Comment      string         `json:"comment"`
	Properties   []Property     `json:"properties"`
	Connectivity []Connectivity `json:"connectivity"`
}

// Schema represents the complete database schema
type Schema struct {
	NodeTables []NodeTable `json:"nodeTables"`
	RelTables  []RelTable  `json:"relTables"`
}

// KuzuDB wrapper
type KuzuDB struct {
	db   *kuzu.Database
	conn *kuzu.Connection
}

func NewKuzuDB(dbPath string, readOnly bool) (*KuzuDB, error) {
	// Create system config
	systemConfig := kuzu.DefaultSystemConfig()
	systemConfig.ReadOnly = readOnly

	// Open database
	db, err := kuzu.OpenDatabase(dbPath, systemConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Open connection
	conn, err := kuzu.OpenConnection(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	return &KuzuDB{
		db:   db,
		conn: conn,
	}, nil
}

func (k *KuzuDB) Close() {
	if k.conn != nil {
		k.conn.Close()
	}
	if k.db != nil {
		k.db.Close()
	}
}

// convertBigInt handles bigint conversion like JS bigIntReplacer
func convertBigInt(value any) any {
	switch v := value.(type) {
	case int64:
		return strconv.FormatInt(v, 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	default:
		return value
	}
}

// executeQueryAndGetAll mimics JS queryResult.getAll()
func (k *KuzuDB) executeQueryAndGetAll(query string) ([]map[string]interface{}, error) {
	queryResult, err := k.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer queryResult.Close()

	var results []map[string]interface{}

	for queryResult.HasNext() {
		tuple, err := queryResult.Next()
		if err != nil {
			return nil, err
		}

		rowMap, err := tuple.GetAsMap()
		tuple.Close()
		if err != nil {
			return nil, err
		}

		// Handle bigint conversion
		convertedMap := make(map[string]interface{})
		for key, value := range rowMap {
			convertedMap[key] = convertBigInt(value)
		}

		results = append(results, convertedMap)
	}

	return results, nil
}

// GetSchema implements the complete schema retrieval logic from JS
func (k *KuzuDB) GetSchema() (*Schema, error) {
	// Step 1: Get all tables
	tables, err := k.executeQueryAndGetAll("CALL show_tables() RETURN *")
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	var nodeTables []NodeTable
	var relTables []RelTable

	// Step 2: Process each table
	for _, table := range tables {
		tableName, ok := table["name"].(string)
		if !ok {
			continue
		}

		tableType, ok := table["type"].(string)
		if !ok {
			continue
		}

		tableComment := ""
		if comment, ok := table["comment"].(string); ok {
			tableComment = comment
		}

		// Step 3: Get table properties
		properties, err := k.getTableProperties(tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to get properties for table %s: %w", tableName, err)
		}

		// Step 4: Process based on table type
		if tableType == TABLE_TYPE_NODE {
			nodeTable := NodeTable{
				Name:       tableName,
				Comment:    tableComment,
				Properties: properties,
			}
			nodeTables = append(nodeTables, nodeTable)

		} else if tableType == TABLE_TYPE_REL {
			// For rel tables, remove isPrimaryKey from properties
			relProperties := make([]Property, len(properties))
			for i, prop := range properties {
				relProperties[i] = Property{
					Name: prop.Name,
					Type: prop.Type,
					// Note: isPrimaryKey omitted for rel tables
				}
			}

			// Get connectivity information
			connectivity, err := k.getTableConnectivity(tableName)
			if err != nil {
				return nil, fmt.Errorf("failed to get connectivity for table %s: %w", tableName, err)
			}

			relTable := RelTable{
				Name:         tableName,
				Comment:      tableComment,
				Properties:   relProperties,
				Connectivity: connectivity,
			}
			relTables = append(relTables, relTable)
		}
	}

	// Step 5: Sort tables alphabetically (like JS version)
	sort.Slice(nodeTables, func(i, j int) bool {
		return nodeTables[i].Name < nodeTables[j].Name
	})
	sort.Slice(relTables, func(i, j int) bool {
		return relTables[i].Name < relTables[j].Name
	})

	return &Schema{
		NodeTables: nodeTables,
		RelTables:  relTables,
	}, nil
}

// getTableProperties gets properties for a specific table
func (k *KuzuDB) getTableProperties(tableName string) ([]Property, error) {
	query := fmt.Sprintf("CALL TABLE_INFO('%s') RETURN *", tableName)
	propertyRows, err := k.executeQueryAndGetAll(query)
	if err != nil {
		return nil, err
	}

	var properties []Property
	for _, row := range propertyRows {
		property := Property{}

		if name, ok := row["name"].(string); ok {
			property.Name = name
		}

		if propType, ok := row["type"].(string); ok {
			property.Type = propType
		}

		// Handle primary key field (might be bool or string)
		if pk, ok := row["primary key"]; ok {
			switch pkVal := pk.(type) {
			case bool:
				property.IsPrimaryKey = pkVal
			case string:
				property.IsPrimaryKey = (pkVal == "true" || pkVal == "True")
			}
		}

		properties = append(properties, property)
	}

	return properties, nil
}

// getTableConnectivity gets connectivity info for relationship tables
func (k *KuzuDB) getTableConnectivity(tableName string) ([]Connectivity, error) {
	query := fmt.Sprintf("CALL SHOW_CONNECTION('%s') RETURN *", tableName)
	connectivityRows, err := k.executeQueryAndGetAll(query)
	if err != nil {
		return nil, err
	}

	var connectivity []Connectivity
	for _, row := range connectivityRows {
		conn := Connectivity{}

		if src, ok := row["source table name"].(string); ok {
			conn.Src = src
		}

		if dst, ok := row["destination table name"].(string); ok {
			conn.Dst = dst
		}

		connectivity = append(connectivity, conn)
	}

	return connectivity, nil
}

// ExecuteQueryAndGetAll executes a Cypher query and returns all results
func (k *KuzuDB) ExecuteQueryAndGetAll(cypher string) ([]map[string]interface{}, error) {
	return k.executeQueryAndGetAll(cypher)
}

// Alternative method if you want more control over query execution
func (k *KuzuDB) ExecuteQuery(cypher string) (*QueryResult, error) {
	queryResult, err := k.conn.Query(cypher)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	return &QueryResult{
		result: queryResult,
		db:     k,
	}, nil
}

// QueryResult wraps Kuzu QueryResult for easier handling
type QueryResult struct {
	result *kuzu.QueryResult
	db     *KuzuDB
}

// GetAll returns all rows as maps (like JS getAll())
func (qr *QueryResult) GetAll() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	for qr.result.HasNext() {
		tuple, err := qr.result.Next()
		if err != nil {
			return nil, err
		}

		rowMap, err := tuple.GetAsMap()
		tuple.Close()
		if err != nil {
			return nil, err
		}

		// Handle bigint conversion
		convertedMap := make(map[string]interface{})
		for key, value := range rowMap {
			convertedMap[key] = convertBigInt(value)
		}

		results = append(results, convertedMap)
	}

	return results, nil
}

// Close closes the query result
func (qr *QueryResult) Close() {
	if qr.result != nil {
		qr.result.Close()
	}
}

// HasNext checks if there are more results
func (qr *QueryResult) HasNext() bool {
	return qr.result.HasNext()
}

// Next returns the next tuple
func (qr *QueryResult) Next() (map[string]interface{}, error) {
	if !qr.result.HasNext() {
		return nil, fmt.Errorf("no more results")
	}

	tuple, err := qr.result.Next()
	if err != nil {
		return nil, err
	}
	defer tuple.Close()

	rowMap, err := tuple.GetAsMap()
	if err != nil {
		return nil, err
	}

	// Handle bigint conversion
	convertedMap := make(map[string]interface{})
	for key, value := range rowMap {
		convertedMap[key] = convertBigInt(value)
	}

	return convertedMap, nil
}

// GetColumnNames returns column names
func (qr *QueryResult) GetColumnNames() []string {
	return qr.result.GetColumnNames()
}

// ToString returns string representation
func (qr *QueryResult) ToString() string {
	return qr.result.ToString()
}
