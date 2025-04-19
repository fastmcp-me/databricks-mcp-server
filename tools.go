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

func listAllTables(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments
	catalogName := arguments["catalog"].(string)
	schemaName := arguments["schema"].(string)
	pattern, ok := arguments["filter_pattern"]
	if !ok {
		pattern = ""
	}
	tables, err := w.Tables.ListAll(ctx, catalog.ListTablesRequest{
		CatalogName: catalogName,
		SchemaName:  schemaName,
	})
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Error listing tables", err), nil
	}
	if pattern != "" {
		tables, err = filterTables(tables, pattern.(string))
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Error filtering tables using pattern", err), nil
		}
	}
	res, err := json.Marshal(tables)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Error marshalling tables into JSON", err), nil
	}
	return mcp.NewToolResultText(string(res)), nil
}

func executeSQL(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	server := server.ServerFromContext(ctx)
	arguments := request.Params.Arguments
	statement := arguments["statement"].(string)
	timeout, _ := arguments["timeout_seconds"].(float64)
	warehouses, err := w.Warehouses.ListAll(ctx, sql.ListWarehousesRequest{})
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Error listing warehouses", err), nil
	}
	if len(warehouses) == 0 {
		return mcp.NewToolResultError("No warehouses available"), nil
	}
	res, err := w.StatementExecution.ExecuteStatement(ctx, sql.ExecuteStatementRequest{
		RowLimit:        100,
		Statement:       statement,
		WaitTimeout:     "10s",
		WarehouseId:     warehouses[0].Id,
		ForceSendFields: nil,
	})
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Error executing SQL statement", err), nil
	}
	var i = 0
	for (i < 2) && (res.Status.State == sql.StatementStateFailed || res.Status.State == sql.StatementStateCanceled || res.Status.State == sql.StatementStatePending) {
		err = server.SendNotificationToClient(ctx, "notifications/progress", map[string]interface{}{
			"message":       "The statement is still running, please wait...",
			"progressToken": request.Params.Meta.ProgressToken,
		})
		if err != nil {
			return nil, err
		}
		time.Sleep(time.Duration(timeout/2) * time.Second)
		res, err = w.StatementExecution.GetStatementByStatementId(ctx, res.StatementId)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Error getting statement status", err), nil
		}
		i++
	}
	var resultDataArray [][]string
	if res.Status.Error != nil {
		return mcp.NewToolResultErrorFromErr(fmt.Sprintf("Error executing the statement, current status %s", res.Status.State), fmt.Errorf(res.Status.Error.Message)), nil
	}
	if res.Status.State != sql.StatementStateSucceeded {
		return mcp.NewToolResultError(fmt.Sprintf("Error executing the statement, current status %s", res.Status.State)), nil
	}
	resultData := res.Result
	resultDataArray = append(resultDataArray, resultData.DataArray...)
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
	response, err := json.Marshal(map[string]interface{}{
		"columns": res.Manifest.Schema.Columns,
		"rows":    resultDataArray,
	})
	if err != nil {
		return mcp.NewToolResultErrorFromErr("Error marshalling statement result into JSON", err), nil
	}
	return mcp.NewToolResultText(string(response)), nil
}
