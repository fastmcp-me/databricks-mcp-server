package tools

import (
	"context"
	"regexp"

	"github.com/databricks/databricks-sdk-go/listing"
	"github.com/databricks/databricks-sdk-go/service/catalog"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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
func ListTables() server.ToolHandlerFunc {
	return ExecuteOperation(func(ctx context.Context, request mcp.CallToolRequest) (interface{}, error) {
		w, err := WorkspaceClientFromContext(ctx)
		if err != nil {
			return nil, err
		}

		catalogName := ExtractStringParam(request, "catalog", "")
		schemaName := ExtractStringParam(request, "schema", "")
		tableNamePattern := ExtractStringParam(request, "table_name_pattern", ".*")
		omitProperties := ExtractBoolParam(request, "omit_properties", true)
		omitColumns := ExtractBoolParam(request, "omit_columns", false)
		maxResults := ExtractIntParam(request, "max_results", 10)

		// Retrieve all tables from the specified catalog and schema
		tablesIt := w.Tables.List(ctx, catalog.ListTablesRequest{
			CatalogName:    catalogName,
			SchemaName:     schemaName,
			OmitProperties: omitProperties,
			OmitColumns:    omitColumns,
			MaxResults:     maxResults + 1,
		})
		tables, err := listing.ToSliceN[catalog.TableInfo](ctx, tablesIt, maxResults)
		if err != nil {
			return nil, err
		}

		var truncated = false
		if len(tables) > maxResults {
			tables = tables[:maxResults]
			truncated = true
		}

		// Apply filter if pattern is not ".*" (match everything)
		if tableNamePattern != "" && tableNamePattern != ".*" {
			tables, err = filterTables(tables, tableNamePattern)
			if err != nil {
				return nil, err
			}
		}

		// Return a structured response
		return map[string]interface{}{
			"tables":      tables,
			"total_count": len(tables),
			"truncated":   truncated,
		}, nil
	})
}
