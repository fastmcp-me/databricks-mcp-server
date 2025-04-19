package databricks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// MCPProvider implements the MCP interface for Databricks
type MCPProvider struct {
	client *Client
	config *Config
	server *http.Server
}

// Config holds the configuration for the Databricks MCP provider
type Config struct {
	BaseURL      string `json:"base_url"`
	Token        string `json:"token"`
	PollInterval int    `json:"poll_interval"`
	ServerPort   int    `json:"server_port"`
}

// NewMCPProvider creates a new Databricks MCP provider
func NewMCPProvider(config *Config) (*MCPProvider, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base_url is required")
	}
	if config.Token == "" {
		return nil, fmt.Errorf("token is required")
	}
	if config.PollInterval == 0 {
		config.PollInterval = 60 // Default to 60 seconds
	}
	if config.ServerPort == 0 {
		config.ServerPort = 8080 // Default to port 8080
	}

	client, err := NewClient(config.BaseURL, config.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &MCPProvider{
		client: client,
		config: config,
	}, nil
}

// GetResources retrieves all resources from Databricks
func (p *MCPProvider) GetResources(ctx context.Context) ([]map[string]interface{}, error) {
	// Get clusters
	clusters, err := p.client.ListClusters(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	// Get jobs
	jobs, err := p.client.ListJobs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	// Combine resources
	resources := make([]map[string]interface{}, 0, len(clusters)+len(jobs))

	// Add clusters
	for _, cluster := range clusters {
		resource := map[string]interface{}{
			"id":         cluster["cluster_id"],
			"name":       cluster["cluster_name"],
			"type":       "cluster",
			"properties": cluster,
		}
		resources = append(resources, resource)
	}

	// Add jobs
	for _, job := range jobs {
		resource := map[string]interface{}{
			"id":         job["job_id"],
			"name":       job["settings"].(map[string]interface{})["name"],
			"type":       "job",
			"properties": job,
		}
		resources = append(resources, resource)
	}

	return resources, nil
}

// GetResource retrieves a specific resource from Databricks
func (p *MCPProvider) GetResource(ctx context.Context, id string, resourceType string) (map[string]interface{}, error) {
	if resourceType == "cluster" {
		cluster, err := p.client.GetCluster(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get cluster: %w", err)
		}

		resource := map[string]interface{}{
			"id":         cluster["cluster_id"],
			"name":       cluster["cluster_name"],
			"type":       "cluster",
			"properties": cluster,
		}

		return resource, nil
	}

	// Add support for other resource types as needed
	return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
}

// StartPolling starts polling for resource changes
func (p *MCPProvider) StartPolling(ctx context.Context, callback func([]map[string]interface{})) error {
	ticker := time.NewTicker(time.Duration(p.config.PollInterval) * time.Second)
	defer ticker.Stop()

	// Initial fetch
	resources, err := p.GetResources(ctx)
	if err != nil {
		return fmt.Errorf("failed to get initial resources: %w", err)
	}
	callback(resources)

	// Start polling
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			resources, err := p.GetResources(ctx)
			if err != nil {
				log.Printf("Error fetching resources: %v", err)
				continue
			}
			callback(resources)
		}
	}
}

// StartServer starts the MCP server to serve context for LLMs
func (p *MCPProvider) StartServer(ctx context.Context) error {
	router := mux.NewRouter()

	// Define API endpoints
	router.HandleFunc("/api/resources", p.handleGetResources).Methods("GET")
	router.HandleFunc("/api/resources/{type}/{id}", p.handleGetResource).Methods("GET")
	router.HandleFunc("/api/context", p.handleGetContext).Methods("GET")

	// Create server
	p.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", p.config.ServerPort),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting MCP server on port %d", p.config.ServerPort)
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for context cancellation to stop server
	<-ctx.Done()
	log.Println("Shutting down server...")

	// Create a deadline for server shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server
	if err := p.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Println("Server stopped")
	return nil
}

// handleGetResources handles requests to get all resources
func (p *MCPProvider) handleGetResources(w http.ResponseWriter, r *http.Request) {
	resources, err := p.GetResources(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting resources: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}

// handleGetResource handles requests to get a specific resource
func (p *MCPProvider) handleGetResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	resourceType := vars["type"]
	resourceID := vars["id"]

	resource, err := p.GetResource(r.Context(), resourceID, resourceType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting resource: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resource)
}

// handleGetContext handles requests to get context for LLMs
func (p *MCPProvider) handleGetContext(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	// Get resources
	resources, err := p.GetResources(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting resources: %v", err), http.StatusInternalServerError)
		return
	}

	// Filter resources based on query (simple contains check for demo)
	var filteredResources []map[string]interface{}
	for _, resource := range resources {
		name, ok := resource["name"].(string)
		if ok && contains(name, query) {
			filteredResources = append(filteredResources, resource)
		}
	}

	// Create context response
	context := map[string]interface{}{
		"query":     query,
		"resources": filteredResources,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(context)
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return strings.Contains(s, substr)
}

// LoadConfig loads the configuration from a JSON file
func LoadConfig(configData []byte) (*Config, error) {
	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &config, nil
}
