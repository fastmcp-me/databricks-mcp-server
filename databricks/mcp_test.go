package databricks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewMCPProvider(t *testing.T) {
	// Test with valid config
	config := &Config{
		BaseURL: "https://databricks.example.com",
		Token:   "test-token",
	}

	provider, err := NewMCPProvider(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if provider.client == nil {
		t.Error("Expected client to be initialized")
	}

	if provider.config != config {
		t.Error("Expected config to be the same instance")
	}

	// Test with default poll interval
	if provider.config.PollInterval != 60 {
		t.Errorf("Expected default PollInterval to be 60, got %d", provider.config.PollInterval)
	}

	// Test with missing BaseURL
	config = &Config{
		Token: "test-token",
	}

	_, err = NewMCPProvider(config)
	if err == nil {
		t.Error("Expected error for missing BaseURL, got nil")
	}

	// Test with missing Token
	config = &Config{
		BaseURL: "https://databricks.example.com",
	}

	_, err = NewMCPProvider(config)
	if err == nil {
		t.Error("Expected error for missing Token, got nil")
	}
}

func TestGetResources(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/2.0/clusters/list" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"clusters": [{"cluster_id": "test-cluster-1", "cluster_name": "Test Cluster 1"}, {"cluster_id": "test-cluster-2", "cluster_name": "Test Cluster 2"}]}`))
		} else {
			t.Errorf("Unexpected request to %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Token:   "test-token",
	}

	provider, err := NewMCPProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	resources, err := provider.GetResources(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(resources) != 2 {
		t.Fatalf("Expected 2 resources, got %d", len(resources))
	}

	if resources[0]["id"] != "test-cluster-1" {
		t.Errorf("Expected first resource id to be 'test-cluster-1', got %v", resources[0]["id"])
	}

	if resources[0]["name"] != "Test Cluster 1" {
		t.Errorf("Expected first resource name to be 'Test Cluster 1', got %v", resources[0]["name"])
	}

	if resources[0]["type"] != "cluster" {
		t.Errorf("Expected first resource type to be 'cluster', got %v", resources[0]["type"])
	}
}

func TestGetResource(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/2.0/clusters/get" && r.URL.Query().Get("cluster_id") == "test-cluster" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"cluster_id": "test-cluster", "cluster_name": "Test Cluster"}`))
		} else {
			t.Errorf("Unexpected request to %s with query %v", r.URL.Path, r.URL.Query())
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := &Config{
		BaseURL: server.URL,
		Token:   "test-token",
	}

	provider, err := NewMCPProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	resource, err := provider.GetResource(context.Background(), "test-cluster")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resource["id"] != "test-cluster" {
		t.Errorf("Expected resource id to be 'test-cluster', got %v", resource["id"])
	}

	if resource["name"] != "Test Cluster" {
		t.Errorf("Expected resource name to be 'Test Cluster', got %v", resource["name"])
	}

	if resource["type"] != "cluster" {
		t.Errorf("Expected resource type to be 'cluster', got %v", resource["type"])
	}
}

func TestLoadConfig(t *testing.T) {
	configJSON := []byte(`{
		"base_url": "https://databricks.example.com",
		"token": "test-token",
		"poll_interval": 30
	}`)

	config, err := LoadConfig(configJSON)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if config.BaseURL != "https://databricks.example.com" {
		t.Errorf("Expected BaseURL to be 'https://databricks.example.com', got %s", config.BaseURL)
	}

	if config.Token != "test-token" {
		t.Errorf("Expected Token to be 'test-token', got %s", config.Token)
	}

	if config.PollInterval != 30 {
		t.Errorf("Expected PollInterval to be 30, got %d", config.PollInterval)
	}

	// Test with invalid JSON
	_, err = LoadConfig([]byte(`{invalid json`))
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestStartPolling(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/2.0/clusters/list" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"clusters": [{"cluster_id": "test-cluster", "cluster_name": "Test Cluster"}]}`))
		} else {
			t.Errorf("Unexpected request to %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := &Config{
		BaseURL:      server.URL,
		Token:        "test-token",
		PollInterval: 1, // 1 second for faster testing
	}

	provider, err := NewMCPProvider(config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create a channel to receive resources
	resourcesCh := make(chan []map[string]interface{}, 2)

	// Start polling in a goroutine
	go func() {
		err := provider.StartPolling(ctx, func(resources []map[string]interface{}) {
			resourcesCh <- resources
		})
		if err != nil && err != context.DeadlineExceeded {
			t.Errorf("Unexpected error: %v", err)
		}
	}()

	// Wait for at least one resource update
	select {
	case resources := <-resourcesCh:
		if len(resources) != 1 {
			t.Errorf("Expected 1 resource, got %d", len(resources))
		}
		if resources[0]["id"] != "test-cluster" {
			t.Errorf("Expected resource id to be 'test-cluster', got %v", resources[0]["id"])
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for resources")
	}
}
