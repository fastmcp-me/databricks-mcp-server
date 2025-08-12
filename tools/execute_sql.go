package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/databricks/databricks-sdk-go/service/sql"
	"github.com/mark3labs/mcp-go/mcp"
)

// ExecuteSQL executes a SQL statement on a Databricks warehouse and returns the results.
// It handles statement execution, polling for completion, and fetching result chunks.
func ExecuteSQL(ctx context.Context, request mcp.CallToolRequest) (interface{}, error) {
	w, err := WorkspaceClientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	sqlStatement := request.GetString("statement", "")
	timeoutSeconds := request.GetFloat("execution_timeout_seconds", 60)
	maxRows := request.GetInt("max_rows", 100)
	warehouseId := request.GetString("warehouse_id", "")

	// Convert timeout to string format for API and calculate a polling interval
	pollingInterval := 10 * time.Second
	// Poll for statement completion
	maxAttempts := int(timeoutSeconds / 10)

	// Determine which warehouse to use
	if warehouseId == "" {
		// Get available warehouses and use the first one
		warehouses, err := w.Warehouses.ListAll(ctx, sql.ListWarehousesRequest{})
		if err != nil {
			return nil, fmt.Errorf("error listing warehouses: %w", err)
		}
		if len(warehouses) == 0 {
			return nil, fmt.Errorf("no warehouses available")
		}
		warehouseId = warehouses[0].Id
	}

	// Execute the SQL statement with the specified row limit
	res, err := w.StatementExecution.ExecuteStatement(ctx, sql.ExecuteStatementRequest{
		RowLimit:    int64(maxRows),
		Statement:   sqlStatement,
		WaitTimeout: "5s",
		WarehouseId: warehouseId,
	})
	if err != nil {
		return nil, fmt.Errorf("error executing SQL statement: %w", err)
	}

	attempts := 0

	for attempts < maxAttempts && res.Status.State != sql.StatementStateSucceeded && res.Status.Error == nil {
		// Send progress notification
		err = SendProgressNotification(ctx,
			fmt.Sprintf("Statement execution in progress (%d seconds), current status: %s", attempts*10, res.Status.State),
			attempts, maxAttempts)
		if err != nil {
			return nil, err
		}

		// Wait before checking again
		time.Sleep(pollingInterval)

		// Check statement status
		res, err = w.StatementExecution.GetStatementByStatementId(ctx, res.StatementId)
		if err != nil {
			return nil, fmt.Errorf("error getting statement status: %w", err)
		}
		attempts++
	}

	// Handle statement errors
	if res.Status.Error != nil {
		return nil, fmt.Errorf("error executing the statement, current status %s: %s",
			res.Status.State, res.Status.Error.Message)
	}

	if res.Status.State != sql.StatementStateSucceeded {
		_ = w.StatementExecution.CancelExecution(ctx, sql.CancelExecutionRequest{
			StatementId: res.StatementId,
		})
		return nil, fmt.Errorf("statement execution timed out after %v seconds, current status %s.\nHint: Try with a higher timeout or simplying the query.", timeoutSeconds, res.Status.State)
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
			return nil, fmt.Errorf("error getting statement result chunk: %w", err)
		}
		resultDataArray = append(resultDataArray, resultData.DataArray...)
	}

	// Return structured results
	return map[string]interface{}{
		"columns": res.Manifest.Schema.Columns,
		"rows":    resultDataArray,
	}, nil
}
