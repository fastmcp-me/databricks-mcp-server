# Databricks MCP Server

A Model Control Protocol (MCP) server for interacting with Databricks.

## Features

The Databricks MCP Server provides a Model Control Protocol (MCP) interface to interact with Databricks workspaces. It offers the following functionalities:

### List Catalogs

Lists all catalogs available in the Databricks workspace.

**Tool name:** `list_catalogs`

**Parameters:** None

**Returns:** JSON array of catalog objects

### List Schemas

Lists all schemas in a specified catalog.

**Tool name:** `list_schemas`

**Parameters:**
- `catalog` (string, required): The name of the catalog to list schemas from

**Returns:** JSON array of schema objects

### List Tables

Lists all tables in a specified schema, with optional filtering by name pattern.

**Tool name:** `list_tables`

**Parameters:**
- `catalog` (string, required): The name of the catalog
- `schema` (string, required): The name of the schema
- `filter_pattern` (string, optional, default: ".*"): Regular expression pattern to filter table names

**Returns:** JSON array of table objects

### Execute SQL Statements

Executes SQL statements on a Databricks warehouse and returns the results.

**Tool name:** `execute_sql_statement`

**Parameters:**
- `statement` (string, required): The SQL statement to execute
- `timeout_seconds` (number, optional, default: 60): Timeout in seconds for the statement execution
- `row_limit` (number, optional, default: 100): Maximum number of rows to return in the result

**Returns:** JSON object containing columns and rows from the query result

## Installation

You can download the latest release for your platform from the [Releases](https://github.com/yourusername/DatabricksMCP/releases) page.

### Supported Platforms

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
./DatabricksMCP
```

The server will start and listen for MCP protocol commands on standard input/output.

## Development

### Prerequisites

- Go 1.24 or later
- Databricks account for testing
