package main

import (
	"fmt"

	"github.com/databricks/databricks-sdk-go"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var w *databricks.WorkspaceClient

func init() {
	w = databricks.Must(databricks.NewWorkspaceClient())
}

func main() {
	// Create an MCP server
	s := server.NewMCPServer(
		"Databricks MCP Server 🚀",
		"1.1.0",
	)

	// Add tool handler
	s.AddTool(mcp.NewTool("list_catalogs",
		mcp.WithDescription("List all catalogs in the databricks workspace"),
	), listAllCatalogs)

	s.AddTool(mcp.NewTool("list_schemas",
		mcp.WithDescription("List all schemas in a catalog"),
		mcp.WithString("catalog", mcp.Description("Catalog name"), mcp.Required()),
	), listAllSchemas)

	s.AddTool(mcp.NewTool("list_tables",
		mcp.WithDescription("List all tables in a schema"),
		mcp.WithString("catalog", mcp.Description("Catalog name"), mcp.Required()),
		mcp.WithString("schema", mcp.Description("Schema name"), mcp.Required()),
		mcp.WithString("filter_pattern", mcp.Description("Pattern to filter tables, expect a valid regex"), mcp.DefaultString(".*")),
	), listAllTables)
	s.AddTool(mcp.NewTool("execute_sql_statement",
		mcp.WithDescription("Execute SQL statement"),
		mcp.WithString("statement", mcp.Description("SQL statement to execute"), mcp.Required()),
		mcp.WithNumber("timeout_seconds", mcp.Description("Timeout in seconds"), mcp.DefaultNumber(60)),
	), executeSQL)
	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
