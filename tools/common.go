package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/databricks/databricks-sdk-go"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

// workspaceClientKey is the key used to store the workspace client in the context
const workspaceClientKey contextKey = "workspaceClient"

// WithWorkspaceClient returns a new context with the workspace client added
func WithWorkspaceClient(ctx context.Context, w *databricks.WorkspaceClient) context.Context {
	return context.WithValue(ctx, workspaceClientKey, w)
}

// WorkspaceClientFromContext retrieves the workspace client from the context
func WorkspaceClientFromContext(ctx context.Context) (*databricks.WorkspaceClient, error) {
	w, ok := ctx.Value(workspaceClientKey).(*databricks.WorkspaceClient)
	if !ok || w == nil {
		return nil, fmt.Errorf("workspace client not found in context")
	}
	return w, nil
}

// DatabricksOperation represents a generic operation on Databricks
type DatabricksOperation func(ctx context.Context, request mcp.CallToolRequest) (interface{}, error)

// ExecuteOperation is a helper function that executes a Databricks operation and handles common error patterns
func ExecuteOperation(operation DatabricksOperation) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := operation(ctx, request)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Error executing operation", err), nil
		}

		// Marshal the result to JSON
		jsonResult, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Error marshalling result into JSON", err), nil
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	}
}

// WithWorkspaceClientHandler wraps a tool handler function with the workspace client
func WithWorkspaceClientHandler(w *databricks.WorkspaceClient, handler server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Add the workspace client to the context
		ctx = WithWorkspaceClient(ctx, w)
		return handler(ctx, request)
	}
}

// ExtractStringParam extracts a string parameter from the request arguments
func ExtractStringParam(request mcp.CallToolRequest, paramName string, defaultValue string) string {
	if val, ok := request.Params.Arguments[paramName].(string); ok {
		return val
	}
	return defaultValue
}

// ExtractBoolParam extracts a boolean parameter from the request arguments
func ExtractBoolParam(request mcp.CallToolRequest, paramName string, defaultValue bool) bool {
	if val, ok := request.Params.Arguments[paramName].(bool); ok {
		return val
	}
	return defaultValue
}

// ExtractIntParam extracts an integer parameter from the request arguments
func ExtractIntParam(request mcp.CallToolRequest, paramName string, defaultValue int) int {
	if val, ok := request.Params.Arguments[paramName].(float64); ok {
		return int(val)
	}
	return defaultValue
}

// ExtractFloatParam extracts a float parameter from the request arguments
func ExtractFloatParam(request mcp.CallToolRequest, paramName string, defaultValue float64) float64 {
	if val, ok := request.Params.Arguments[paramName].(float64); ok {
		return val
	}
	return defaultValue
}

// SendProgressNotification sends a progress notification to the client
func SendProgressNotification(ctx context.Context, message string, progress, total int) error {
	mcpServer := server.ServerFromContext(ctx)
	if mcpServer == nil {
		return fmt.Errorf("server not found in context")
	}

	var token interface{} = 0
	return mcpServer.SendNotificationToClient(ctx, "notifications/progress", map[string]interface{}{
		"message":       message,
		"progressToken": token,
		"progress":      progress,
		"total":         total,
	})
}
