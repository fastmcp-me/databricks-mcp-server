# DatabricksMCP

DatabricksMCP is a Go implementation of the Model Context Protocol (MCP) for Databricks. It allows you to interact with Databricks resources using the MCP interface and serves context for LLMs.

## Features

- Retrieve information about Databricks clusters and jobs
- Poll for changes to Databricks resources
- Serve context for LLMs via HTTP API
- Authentication using databricks-cli credentials or token
- Easy configuration via JSON

## Installation

```bash
go get github.com/yourusername/DatabricksMCP
```

## Usage

### Basic Usage (Client Mode)

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/DatabricksMCP/databricks"
)

func main() {
	// Create configuration
	config := &databricks.Config{
		BaseURL:      "https://your-databricks-instance.cloud.databricks.com",
		Token:        "your-databricks-token",
		PollInterval: 60, // seconds
	}

	// Create MCP provider
	provider, err := databricks.NewMCPProvider(config)
	if err != nil {
		log.Fatalf("Failed to create MCP provider: %v", err)
	}

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Received shutdown signal, stopping...")
		cancel()
	}()

	// Start polling for resources
	log.Println("Starting to poll for Databricks resources...")
	err = provider.StartPolling(ctx, func(resources []map[string]interface{}) {
		log.Printf("Received %d resources", len(resources))
		// Process resources here
	})

	if err != nil && err != context.Canceled {
		log.Fatalf("Polling failed: %v", err)
	}
}
```

### Server Mode

You can run the application in server mode to serve context for LLMs:

```go
// Start the MCP server
err = provider.StartServer(ctx)
if err != nil && err != context.Canceled {
    log.Fatalf("Server failed: %v", err)
}
```

The server provides the following endpoints:
- `GET /api/resources` - List all resources
- `GET /api/resources/{type}/{id}` - Get a specific resource
- `GET /api/context?query={query}` - Get context for LLMs based on a query

### Using Configuration File

You can also load configuration from a JSON file:

```go
// Load configuration from file
configData, err := ioutil.ReadFile("config.json")
if err != nil {
	log.Fatalf("Failed to read config file: %v", err)
}

config, err := databricks.LoadConfig(configData)
if err != nil {
	log.Fatalf("Failed to parse config: %v", err)
}

// Create MCP provider
provider, err := databricks.NewMCPProvider(config)
```

Example config.json:
```json
{
  "base_url": "https://your-databricks-instance.cloud.databricks.com",
  "token": "your-databricks-token",
  "poll_interval": 60,
  "server_port": 8080
}
```

### Authentication

The application supports two authentication methods:

1. **Token-based authentication**: Provide a token in the configuration.
2. **databricks-cli authentication**: The application will automatically use credentials from databricks-cli if available.

## Running the Example

1. Clone the repository
2. Update the `config.json` file with your Databricks credentials
3. Run the example in client mode:

```bash
go run main.go
```

4. Or run in server mode:

```bash
go run main.go --server
```

## Testing

Run the tests with:

```bash
go test ./...
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [mcp-go](https://github.com/mark3labs/mcp-go) - Foundation package for MCP implementation
- [databricks-sdk-go](https://github.com/databricks/databricks-sdk-go) - Official Databricks SDK for Go
- [mcp-databricks-server](https://github.com/RafaelCartenet/mcp-databricks-server) - Inspiration for this implementation
