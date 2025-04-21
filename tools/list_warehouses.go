package tools

import (
	"context"
	"encoding/json"
	"github.com/mark3labs/mcp-go/server"

	"github.com/databricks/databricks-sdk-go"
	"github.com/databricks/databricks-sdk-go/service/sql"
	"github.com/mark3labs/mcp-go/mcp"
)

// ListWarehouses retrieves all SQL warehouses from the Databricks workspace
// and returns them as a JSON string.
func ListWarehouses(w *databricks.WorkspaceClient) server.ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		warehouses, err := w.Warehouses.ListAll(ctx, sql.ListWarehousesRequest{})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Error listing SQL warehouses", err), nil
		}
		res, err := json.Marshal(warehouses)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Error marshalling SQL warehouses into JSON", err), nil
		}
		return mcp.NewToolResultText(string(res)), nil
	}
}
