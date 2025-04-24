package tools

import (
	"context"

	"github.com/databricks/databricks-sdk-go/service/catalog"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ListSchemas retrieves all schemas in the specified catalog
// and returns them as a JSON string.
func ListSchemas() server.ToolHandlerFunc {
	return ExecuteOperation(func(ctx context.Context, request mcp.CallToolRequest) (interface{}, error) {
		w, err := WorkspaceClientFromContext(ctx)
		if err != nil {
			return nil, err
		}
		catalogName := ExtractStringParam(request, "catalog", "")
		return w.Schemas.ListAll(ctx, catalog.ListSchemasRequest{
			CatalogName: catalogName,
		})
	})
}
