package main

import (
	"fmt"
	"log"
	"os"

	"databricks-mcp-server/tools"
	"github.com/databricks/databricks-sdk-go"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// w is the Databricks workspace client used for all API operations
var w *databricks.WorkspaceClient

func init() {
	var err error
	w, err = databricks.NewWorkspaceClient()
	if err != nil {
		log.Fatalf("Failed to initialize Databricks client: %v", err)
	}
}

func main() {
	// Create an MCP server
	s := server.NewMCPServer(
		"Databricks MCP Server",
		Version,
		server.WithLogging(),
	)

	// Add tool handlers for Databricks operations
	s.AddTool(mcp.NewTool("list_catalogs",
		mcp.WithDescription("Lists all catalogs available in the Databricks workspace"),
	), tools.ListCatalogs(w))

	s.AddTool(mcp.NewTool("list_schemas",
		mcp.WithDescription("Lists all schemas in a specified Databricks catalog"),
		mcp.WithString("catalog", mcp.Description("Name of the catalog to list schemas from"), mcp.Required()),
	), tools.ListSchemas(w))

	s.AddTool(mcp.NewTool("list_tables",
		mcp.WithDescription("Lists all tables in a specified Databricks schema with optional filtering"),
		mcp.WithString("catalog", mcp.Description("Name of the catalog containing the schema"), mcp.Required()),
		mcp.WithString("schema", mcp.Description("Name of the schema to list tables from"), mcp.Required()),
		mcp.WithString("table_name_pattern", mcp.Description("Regular expression pattern to filter table names"), mcp.DefaultString(".*")),
		mcp.WithBoolean("omit_properties", mcp.Description("Whether to omit table properties in the response, helps to reduce response size"), mcp.DefaultBool(true)),
		mcp.WithBoolean("omit_columns", mcp.Description("Whether to omit column details in the response"), mcp.DefaultBool(false)),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of tables to return (0 for all, non-recommended)"), mcp.DefaultNumber(10)),
	), tools.ListTables(w))

	s.AddTool(mcp.NewTool("get_table",
		mcp.WithDescription("Gets detailed information about a single Databricks table"),
		mcp.WithString("full_name", mcp.Description("Full name of the table in format 'catalog.schema.table'"), mcp.Required()),
	), tools.GetTable(w))

	s.AddTool(mcp.NewTool("execute_sql",
		mcp.WithDescription("Executes SQL statements on a Databricks warehouse and returns the results"),
		mcp.WithString("statement", mcp.Description("SQL statement to execute"), mcp.Required()),
		mcp.WithNumber("execution_timeout_seconds", mcp.Description("Maximum time in seconds to wait for query execution"), mcp.DefaultNumber(60)),
		mcp.WithNumber("max_rows", mcp.Description("Maximum number of rows to return in the result"), mcp.DefaultNumber(100)),
	), tools.ExecuteSQL(w))

	s.AddTool(mcp.NewTool("list_warehouses",
		mcp.WithDescription("Lists all SQL warehouses available in the Databricks workspace"),
	), tools.ListWarehouses(w))

	// Start the stdio server
	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)
	if err := server.ServeStdio(s, server.WithErrorLogger(logger)); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
