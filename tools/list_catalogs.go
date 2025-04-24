package tools

import (
	"context"

	"github.com/databricks/databricks-sdk-go/service/catalog"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ListCatalogs retrieves all catalogs from the Databricks workspace
// and returns them as a JSON string.
func ListCatalogs() server.ToolHandlerFunc {
	return ExecuteOperation(func(ctx context.Context, _ mcp.CallToolRequest) (interface{}, error) {
		w, err := WorkspaceClientFromContext(ctx)
		if err != nil {
			return nil, err
		}
		return w.Catalogs.ListAll(ctx, catalog.ListCatalogsRequest{})
	})
}
