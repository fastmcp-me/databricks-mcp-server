package tools

import (
	"context"
	"encoding/json"

	"github.com/databricks/databricks-sdk-go"
	"github.com/databricks/databricks-sdk-go/service/catalog"
	"github.com/mark3labs/mcp-go/mcp"
)

// ListSchemas retrieves all schemas in the specified catalog
// and returns them as a JSON string.
func ListSchemas(w *databricks.WorkspaceClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
}
