package databricks

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	baseURL := "https://databricks.example.com"
	token := "test-token"

	client := NewClient(baseURL, token)

	if client.BaseURL != baseURL {
		t.Errorf("Expected BaseURL to be %s, got %s", baseURL, client.BaseURL)
	}

	if client.Token != token {
		t.Errorf("Expected Token to be %s, got %s", token, client.Token)
	}

	if client.HTTPClient == nil {
		t.Error("Expected HTTPClient to be initialized")
	}
}

func TestRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request headers
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header to be 'Bearer test-token', got %s", r.Header.Get("Authorization"))
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header to be 'application/json', got %s", r.Header.Get("Content-Type"))
		}

		// Return a test response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	// Test GET request
	respBody, err := client.Request("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := `{"success": true}`
	if string(respBody) != expected {
		t.Errorf("Expected response body to be %s, got %s", expected, string(respBody))
	}
}

func TestGetCluster(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request path
		if r.URL.Path != "/api/2.0/clusters/get" {
			t.Errorf("Expected path to be '/api/2.0/clusters/get', got %s", r.URL.Path)
		}

		// Check query parameters
		if r.URL.Query().Get("cluster_id") != "test-cluster" {
			t.Errorf("Expected cluster_id to be 'test-cluster', got %s", r.URL.Query().Get("cluster_id"))
		}

		// Return a test response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"cluster_id": "test-cluster", "cluster_name": "Test Cluster"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	// Test GetCluster
	cluster, err := client.GetCluster("test-cluster")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cluster["cluster_id"] != "test-cluster" {
		t.Errorf("Expected cluster_id to be 'test-cluster', got %v", cluster["cluster_id"])
	}

	if cluster["cluster_name"] != "Test Cluster" {
		t.Errorf("Expected cluster_name to be 'Test Cluster', got %v", cluster["cluster_name"])
	}
}

func TestListClusters(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request path
		if r.URL.Path != "/api/2.0/clusters/list" {
			t.Errorf("Expected path to be '/api/2.0/clusters/list', got %s", r.URL.Path)
		}

		// Return a test response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"clusters": [{"cluster_id": "test-cluster-1", "cluster_name": "Test Cluster 1"}, {"cluster_id": "test-cluster-2", "cluster_name": "Test Cluster 2"}]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	// Test ListClusters
	clusters, err := client.ListClusters()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(clusters) != 2 {
		t.Fatalf("Expected 2 clusters, got %d", len(clusters))
	}

	if clusters[0]["cluster_id"] != "test-cluster-1" {
		t.Errorf("Expected first cluster_id to be 'test-cluster-1', got %v", clusters[0]["cluster_id"])
	}

	if clusters[1]["cluster_name"] != "Test Cluster 2" {
		t.Errorf("Expected second cluster_name to be 'Test Cluster 2', got %v", clusters[1]["cluster_name"])
	}
}
