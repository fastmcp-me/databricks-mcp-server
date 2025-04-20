package tools

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/databricks/databricks-sdk-go"
	"github.com/databricks/databricks-sdk-go/service/catalog"
	"github.com/mark3labs/mcp-go/mcp"
)

// filterTables filters a list of tables based on a regex pattern applied to table names.
// Returns the filtered list of tables and any error that occurred during pattern compilation.
func filterTables(tables []catalog.TableInfo, pattern string) ([]catalog.TableInfo, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	var filteredTables []catalog.TableInfo
	for _, table := range tables {
		if regex.MatchString(table.Name) {
			filteredTables = append(filteredTables, table)
		}
	}
	return filteredTables, nil
}

// ListAllTables retrieves all tables in the specified catalog and schema,
// optionally filtering them by a regex pattern, and returns them as a JSON string.
func ListAllTables(w *databricks.WorkspaceClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.Params.Arguments
		catalogName := arguments["catalog"].(string)
		schemaName := arguments["schema"].(string)

		// Get the filter pattern, which defaults to ".*" in main.go
		filterPattern, _ := arguments["filter_pattern"].(string)

		// Retrieve all tables from the specified catalog and schema
		tables, err := w.Tables.ListAll(ctx, catalog.ListTablesRequest{
			CatalogName: catalogName,
			SchemaName:  schemaName,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Error listing tables", err), nil
		}

		// Apply filter if pattern is not ".*" (match everything)
		if filterPattern != "" && filterPattern != ".*" {
			tables, err = filterTables(tables, filterPattern)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Error filtering tables using pattern", err), nil
			}
		}

		// Marshal the tables to JSON
		res, err := json.Marshal(tables)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Error marshalling tables into JSON", err), nil
		}

		return mcp.NewToolResultText(string(res)), nil
	}
}