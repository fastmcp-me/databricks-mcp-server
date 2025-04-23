package tools

import (
	"context"
	"encoding/json"
	"github.com/databricks/databricks-sdk-go/listing"
	"github.com/mark3labs/mcp-go/server"
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

// ListTables retrieves all tables in the specified catalog and schema,
// optionally filtering them by a regex pattern, and returns them as a JSON string.
// It also supports omitting table properties and column details from the response.
// The max_results parameter limits the number of tables returned (0 for all).
func ListTables(w *databricks.WorkspaceClient) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.Params.Arguments
		catalogName := arguments["catalog"].(string)
		schemaName := arguments["schema"].(string)

		// Get the filter pattern, which defaults to ".*" in main.go
		filterPattern, _ := arguments["filter_pattern"].(string)

		// Get the omit_properties and omit_columns parameters, which default to "false" in main.go
		omitProperties, ok := arguments["omit_properties"].(bool)
		if !ok {
			omitProperties = false
		}
		omitColumns, ok := arguments["omit_columns"].(bool)
		if !ok {
			omitColumns = false
		}

		// Get the max_results parameter, which defaults to 10 in main.go
		maxResults, ok := arguments["max_results"].(int)
		if !ok {
			maxResults = 10
		}

		// Retrieve all tables from the specified catalog and schema
		tablesIt := w.Tables.List(ctx, catalog.ListTablesRequest{
			CatalogName:    catalogName,
			SchemaName:     schemaName,
			OmitProperties: omitProperties,
			OmitColumns:    omitColumns,
			MaxResults:     maxResults + 1,
		})
		tables, err := listing.ToSliceN(ctx, tablesIt, maxResults)

		if err != nil {
			return mcp.NewToolResultErrorFromErr("Error listing tables", err), nil
		}

		var truncated = false
		if len(tables) > maxResults {
			tables = tables[:maxResults]
			truncated = true
		}

		// Apply filter if pattern is not ".*" (match everything)
		if filterPattern != "" && filterPattern != ".*" {
			tables, err = filterTables(tables, filterPattern)
			if err != nil {
				return mcp.NewToolResultErrorFromErr("Error filtering tables using pattern", err), nil
			}
		}

		// Marshal the response to JSON
		res, err := json.Marshal(map[string]interface{}{
			"tables":      tables,
			"total_count": len(tables),
			"truncated":   truncated,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Error marshalling tables into JSON", err), nil
		}

		return mcp.NewToolResultText(string(res)), nil
	}
}
