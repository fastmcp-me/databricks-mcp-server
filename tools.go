package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/databricks/databricks-sdk-go/service/catalog"
	"github.com/databricks/databricks-sdk-go/service/sql"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"regexp"
	"time"
)

// listAllCatalogs retrieves all catalogs from the Databricks workspace
// and returns them as a JSON string.
func listAllCatalogs(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	c, err := w.Catalogs.ListAll(ctx, catalog.ListCatalogsRequest{})
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Error listing catalogs", err), nil
	}
	res, err := json.Marshal(c)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Error marshalling catalogs into JSON", err), nil
	}
	return mcp.NewToolResultText(string(res)), nil
}

// listAllSchemas retrieves all schemas in the specified catalog
// and returns them as a JSON string.
func listAllSchemas(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	catalogName := arguments["catalog"].(string)
	schemas, err := w.Schemas.ListAll(ctx, catalog.ListSchemasRequest{
		CatalogName: catalogName,
	})
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Error listing schemas", err), nil
	}
	res, err := json.Marshal(schemas)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Error marshalling schemas into JSON", err), nil
	}
	return mcp.NewToolResultText(string(res)), nil
}

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

// listAllTables retrieves all tables in the specified catalog and schema,
// optionally filtering them by a regex pattern, and returns them as a JSON string.
func listAllTables(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

// executeSQL executes a SQL statement on a Databricks warehouse and returns the results.
// It handles statement execution, polling for completion, and fetching result chunks.
func executeSQL(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	server := server.ServerFromContext(ctx)
	arguments := request.Params.Arguments
	statement := arguments["statement"].(string)
	timeoutSeconds, _ := arguments["timeout_seconds"].(float64)

	// Convert timeout to string format for API and calculate a polling interval
	pollingInterval := time.Duration(timeoutSeconds/4) * time.Second

	// Get available warehouses
	warehouses, err := w.Warehouses.ListAll(ctx, sql.ListWarehousesRequest{})
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Error listing warehouses", err), nil
	}
	if len(warehouses) == 0 {
		return mcp.NewToolResultError("No warehouses available"), nil
	}

	// Execute the SQL statement
	res, err := w.StatementExecution.ExecuteStatement(ctx, sql.ExecuteStatementRequest{
		RowLimit:    100,
		Statement:   statement,
		WaitTimeout: "10s",
		WarehouseId: warehouses[0].Id,
	})
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Error executing SQL statement", err), nil
	}

	// Poll for statement completion
	maxAttempts := 5 // Increased from 2 to give more time for completion
	attempts := 0

	for attempts < maxAttempts && isStatementInProgress(res.Status.State) {
		// Send progress notification to client
		err = server.SendNotificationToClient(ctx, "notifications/progress", map[string]interface{}{
			"message":       "The statement is still running, please wait...",
			"progressToken": request.Params.Meta.ProgressToken,
		})
		if err != nil {
			return nil, err
		}

		// Wait before checking again
		time.Sleep(pollingInterval)

		// Check statement status
		res, err = w.StatementExecution.GetStatementByStatementId(ctx, res.StatementId)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Error getting statement status", err), nil
		}
		attempts++
	}

	// Handle statement errors
	if res.Status.Error != nil {
		return mcp.NewToolResultErrorFromErr(
			fmt.Sprintf("Error executing the statement, current status %s", res.Status.State),
			fmt.Errorf(res.Status.Error.Message)), nil
	}

	if res.Status.State != sql.StatementStateSucceeded {
		return mcp.NewToolResultError(
			fmt.Sprintf("Error executing the statement, current status %s", res.Status.State)), nil
	}

	// Collect all result chunks
	var resultDataArray [][]string
	resultData := res.Result
	resultDataArray = append(resultDataArray, resultData.DataArray...)

	// Fetch additional chunks if available
	for resultData.NextChunkIndex != 0 {
		resultData, err = w.StatementExecution.GetStatementResultChunkN(ctx, sql.GetStatementResultChunkNRequest{
			ChunkIndex:  resultData.NextChunkIndex,
			StatementId: res.StatementId,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Error getting statement result chunk", err), nil
		}
		resultDataArray = append(resultDataArray, resultData.DataArray...)
	}

	// Format and return results
	response, err := json.Marshal(map[string]interface{}{
		"columns": res.Manifest.Schema.Columns,
		"rows":    resultDataArray,
	})
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Error marshalling statement result into JSON", err), nil
	}

	return mcp.NewToolResultText(string(response)), nil
}

// isStatementInProgress returns true if the statement is still running
func isStatementInProgress(state sql.StatementState) bool {
	return state == sql.StatementStatePending
}
