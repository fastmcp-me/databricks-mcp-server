package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"DatabricksMCP/databricks"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.json", "Path to configuration file")
	serverMode := flag.Bool("server", false, "Run in server mode (default: false)")
	flag.Parse()

	// Load configuration
	configData, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	config, err := databricks.LoadConfig(configData)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
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

	// Use a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Start the server if in server mode
	if *serverMode {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Println("Starting MCP server...")
			if err := provider.StartServer(ctx); err != nil && err != context.Canceled {
				log.Fatalf("Server failed: %v", err)
			}
		}()
	}

	// Start polling for resources
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting to poll for Databricks resources...")
		err := provider.StartPolling(ctx, func(resources []map[string]interface{}) {
			log.Printf("Received %d resources", len(resources))

			// Print resource details (only if not in server mode to avoid log spam)
			if !*serverMode {
				for i, resource := range resources {
					resourceJSON, _ := json.MarshalIndent(resource, "", "  ")
					log.Printf("Resource %d: %s", i+1, string(resourceJSON))
				}
			}
		})

		if err != nil && err != context.Canceled {
			log.Fatalf("Polling failed: %v", err)
		}
	}()

	// Wait for all goroutines to finish
	wg.Wait()
	log.Println("Shutdown complete")
}
