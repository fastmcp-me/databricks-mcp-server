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
	), tools.WithWorkspaceClientHandler(w, tools.ExecuteTool(tools.ListCatalogs)))

	s.AddTool(mcp.NewTool("list_schemas",
		mcp.WithDescription("Lists all schemas in a specified Databricks catalog"),
		mcp.WithString("catalog", mcp.Description("Name of the catalog to list schemas from"), mcp.Required()),
	), tools.WithWorkspaceClientHandler(w, tools.ExecuteTool(tools.ListSchemas)))

	s.AddTool(mcp.NewTool("list_tables",
		mcp.WithDescription("Lists all tables in a specified Databricks schema with optional filtering"),
		mcp.WithString("catalog", mcp.Description("Name of the catalog containing the schema"), mcp.Required()),
		mcp.WithString("schema", mcp.Description("Name of the schema to list tables from"), mcp.Required()),
		mcp.WithString("table_name_pattern", mcp.Description("Regular expression pattern to filter table names"), mcp.DefaultString(".*")),
		mcp.WithBoolean("omit_properties", mcp.Description("Whether to omit table properties in the response, helps to reduce response size"), mcp.DefaultBool(true)),
		mcp.WithBoolean("omit_columns", mcp.Description("Whether to omit column details in the response"), mcp.DefaultBool(false)),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of tables to return (0 for all, non-recommended)"), mcp.DefaultNumber(10)),
	), tools.WithWorkspaceClientHandler(w, tools.ExecuteTool(tools.ListTables)))

	s.AddTool(mcp.NewTool("get_table",
		mcp.WithDescription("Gets detailed information about a single Databricks table"),
		mcp.WithString("full_name", mcp.Description("Full name of the table in format 'catalog.schema.table'"), mcp.Required()),
	), tools.WithWorkspaceClientHandler(w, tools.ExecuteTool(tools.GetTable)))

	s.AddTool(mcp.NewTool("execute_sql",
		mcp.WithDescription(`
<use_case>
  Use this tool to execute SQL statements against a Databricks warehouse and retrieve results in JSON format.
</use_case>

<important_notes>
  The flavor of SQL supported is based on the Databricks SQL engine, which is similar to Apache Spark SQL.
  If asked explicitly to use a specific warehouse, you can use the "list_warehouses" tool to get available warehouses.
  Ensure that the SQL is optimized for performance, especially for large datasets; avoid running statements that do not use partitioning or indexing effectively.
</important_notes>
`),
		mcp.WithString("statement", mcp.Description("SQL statement to execute"), mcp.Required()),
		mcp.WithNumber("execution_timeout_seconds", mcp.Description("Maximum time in seconds to wait for query execution"), mcp.DefaultNumber(60)),
		mcp.WithNumber("max_rows", mcp.Description("Maximum number of rows to return in the result"), mcp.DefaultNumber(100)),
		mcp.WithString("warehouse_id", mcp.Description("ID of the warehouse to use for execution. If not specified, the first available warehouse will be used")),
	), tools.WithWorkspaceClientHandler(w, tools.ExecuteTool(tools.ExecuteSQL)))

	s.AddTool(mcp.NewTool("list_warehouses",
		mcp.WithDescription(`
<use_case>
  Use this tool when asked explicitly to use a specific warehouse for SQL execution.
</use_case>
`),
	), tools.WithWorkspaceClientHandler(w, tools.ExecuteTool(tools.ListWarehouses)))

	// Start the stdio server
	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)
	if err := server.ServeStdio(s, server.WithErrorLogger(logger)); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
