package tools

import (
	"context"

	"github.com/databricks/databricks-sdk-go/service/catalog"
	"github.com/mark3labs/mcp-go/mcp"
)

// GetTable retrieves information about a single table using its full name (catalog.schema.table)
// and returns it as a JSON string.
func GetTable(ctx context.Context, request mcp.CallToolRequest) (interface{}, error) {
	w, err := WorkspaceClientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	fullName := request.GetString("full_name", "")

	// Note: The Get method doesn't support omitProperties and omitColumns parameters
	return w.Tables.Get(ctx, catalog.GetTableRequest{
		FullName: fullName,
	})
}
