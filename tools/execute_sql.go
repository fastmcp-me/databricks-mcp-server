package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/databricks/databricks-sdk-go"
	"github.com/databricks/databricks-sdk-go/service/sql"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// isStatementInProgress returns true if the statement is still running
func isStatementInProgress(state sql.StatementState) bool {
	return state == sql.StatementStatePending
}

// ExecuteSQL executes a SQL statement on a Databricks warehouse and returns the results.
// It handles statement execution, polling for completion, and fetching result chunks.
func ExecuteSQL(w *databricks.WorkspaceClient) ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		server := server.ServerFromContext(ctx)
		arguments := request.Params.Arguments
		statement := arguments["statement"].(string)
		timeoutSeconds, _ := arguments["timeout_seconds"].(float64)

		// Get the row limit parameter, default to 100 if not provided
		rowLimit, ok := arguments["row_limit"].(float64)
		if !ok {
			rowLimit = 100
		}

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

		// Execute the SQL statement with the specified row limit
		res, err := w.StatementExecution.ExecuteStatement(ctx, sql.ExecuteStatementRequest{
			RowLimit:    int64(rowLimit),
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
}
