package tools

import (
	"context"
	"encoding/json"
	"github.com/mark3labs/mcp-go/server"

	"github.com/databricks/databricks-sdk-go"
	"github.com/databricks/databricks-sdk-go/service/catalog"
	"github.com/mark3labs/mcp-go/mcp"
)

// GetTable retrieves information about a single table using its full name (catalog.schema.table)
// and returns it as a JSON string.
func GetTable(w *databricks.WorkspaceClient) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments := request.Params.Arguments
		fullName := arguments["full_name"].(string)

		// Note: The Get method doesn't support omitProperties and omitColumns parameters

		// Retrieve the table from the specified catalog and schema
		table, err := w.Tables.Get(ctx, catalog.GetTableRequest{
			FullName: fullName,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Error getting table", err), nil
		}

		// Marshal the table to JSON
		res, err := json.Marshal(table)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Error marshalling table into JSON", err), nil
		}

		return mcp.NewToolResultText(string(res)), nil
	}
}
