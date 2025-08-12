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

// DatabricksTool represents a generic tool on Databricks
type DatabricksTool func(ctx context.Context, request mcp.CallToolRequest) (interface{}, error)

// ExecuteTool is a helper function that executes a Databricks tool and handles common error patterns
func ExecuteTool(tool DatabricksTool) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := tool(ctx, request)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("Error executing tool", err), nil
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
