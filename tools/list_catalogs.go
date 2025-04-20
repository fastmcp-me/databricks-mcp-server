package tools

import (
	"context"
	"encoding/json"

	"github.com/databricks/databricks-sdk-go"
	"github.com/databricks/databricks-sdk-go/service/catalog"
	"github.com/mark3labs/mcp-go/mcp"
)

// ListCatalogs retrieves all catalogs from the Databricks workspace
// and returns them as a JSON string.
func ListCatalogs(w *databricks.WorkspaceClient) ToolHandlerFunc {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
}
