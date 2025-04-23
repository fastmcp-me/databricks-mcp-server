# Databricks MCP Server

A Model Context Protocol (MCP) server for interacting with Databricks.

## Installation

You can download the latest release for your platform from the [Releases](https://github.com/characat0/databricks-mcp-server/releases) page.

### VS Code

Install the Databricks MCP Server extension in VS Code by pressing the following link:

[<img src="https://img.shields.io/badge/VS_Code-VS_Code?style=flat-square&label=Install%20Server&color=0098FF" alt="Install in VS Code">](https://vscode.dev/redirect?url=vscode%3Amcp%2Finstall%3F%257B%2522name%2522%253A%2522databricks%2522%252C%2522command%2522%253A%2522npx%2522%252C%2522args%2522%253A%255B%2522-y%2522%252C%2522databricks-mcp-server%2540latest%2522%255D%257D)

Alternatively, you can install the extension manually by running the following command:

```shell
# For VS Code
code --add-mcp '{"name":"databricks","command":"npx","args":["databricks-mcp-server@latest"]}'
# For VS Code Insiders
code-insiders --add-mcp '{"name":"databricks","command":"npx","args":["databricks-mcp-server@latest"]}'
```

## Tools

The Databricks MCP Server provides a Model Context Protocol (MCP) interface to interact with Databricks workspaces. It offers the following functionalities:

### List Catalogs

Lists all catalogs available in the Databricks workspace.

**Tool name:** `list_catalogs`

**Parameters:** None

**Returns:** JSON array of catalog objects

### List Schemas

Lists all schemas in a specified Databricks catalog.

**Tool name:** `list_schemas`

**Parameters:**
- `catalog` (string, required): Name of the catalog to list schemas from

**Returns:** JSON array of schema objects

### List Tables

Lists all tables in a specified Databricks schema with optional filtering.

**Tool name:** `list_tables`

**Parameters:**
- `catalog` (string, required): Name of the catalog containing the schema
- `schema` (string, required): Name of the schema to list tables from
- `filter_pattern` (string, optional, default: ".*"): Regular expression pattern to filter table names

**Returns:** JSON array of table objects

### Execute SQL

Executes SQL statements on a Databricks SQL warehouse and returns the results.

**Tool name:** `execute_sql`

**Parameters:**
- `statement` (string, required): SQL statement to execute
- `timeout_seconds` (number, optional, default: 60): Timeout in seconds for the statement execution
- `row_limit` (number, optional, default: 100): Maximum number of rows to return in the result

**Returns:** JSON object containing columns and rows from the query result, with information of the 
SQL warehouse used to execute the statement.

### List SQL Warehouses

Lists all SQL warehouses available in the Databricks workspace.

**Tool name:** `list_warehouses`

**Parameters:** None

**Returns:** JSON array of SQL warehouse objects

## Supported Platforms

- Linux (amd64)
- Windows (amd64)
- macOS (Intel/amd64)
- macOS (Apple Silicon/arm64)

## Usage

### Authentication

The application uses Databricks unified authentication. For details on how to configure authentication, please refer to the [Databricks Authentication documentation](https://docs.databricks.com/en/dev-tools/auth.html).

### Running the Server

Start the MCP server:

```bash
./databricks-mcp-server
```

The server will start and listen for MCP protocol commands on standard input/output.

## Development

### Prerequisites

- Go 1.24 or later
