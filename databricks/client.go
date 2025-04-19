package databricks

import (
	"context"
	"fmt"
	"github.com/databricks/databricks-sdk-go"
	"github.com/databricks/databricks-sdk-go/service/compute"
	"github.com/databricks/databricks-sdk-go/service/jobs"
	"log"
)

// Client represents a Databricks API client using the official SDK
type Client struct {
	workspaceClient *databricks.WorkspaceClient
}

// NewClient creates a new Databricks API client using the official SDK
// It will use the databricks-cli authentication if available, or fall back to the provided token
func NewClient(baseURL, token string) (*Client, error) {
	config := databricks.Config{
		Host:  baseURL,
		Token: token,
	}

	workspaceClient, err := databricks.NewWorkspaceClient(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Databricks client: %w", err)
	}

	return &Client{
		workspaceClient: workspaceClient,
	}, nil
}

// GetCluster retrieves information about a specific cluster
func (c *Client) GetCluster(ctx context.Context, clusterID string) (map[string]interface{}, error) {
	request := compute.GetClusterRequest{
		ClusterId: clusterID,
	}
	cluster, err := c.workspaceClient.Clusters.Get(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster: %w", err)
	}

	// Convert to map for compatibility with existing code
	result := make(map[string]interface{})
	result["cluster_id"] = cluster.ClusterId
	result["cluster_name"] = cluster.ClusterName
	result["state"] = string(cluster.State)
	result["creator_user_name"] = cluster.CreatorUserName
	result["spark_version"] = cluster.SparkVersion
	result["node_type_id"] = cluster.NodeTypeId
	result["num_workers"] = cluster.NumWorkers

	return result, nil
}

// ListClusters retrieves a list of all clusters
func (c *Client) ListClusters(ctx context.Context) ([]map[string]interface{}, error) {
	clusters, err := c.workspaceClient.Clusters.ListAll(ctx, compute.ListClustersRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	// Convert to map for compatibility with existing code
	result := make([]map[string]interface{}, 0, len(clusters))
	for _, cluster := range clusters {
		clusterMap := make(map[string]interface{})
		clusterMap["cluster_id"] = cluster.ClusterId
		clusterMap["cluster_name"] = cluster.ClusterName
		clusterMap["state"] = string(cluster.State)
		clusterMap["creator_user_name"] = cluster.CreatorUserName
		clusterMap["spark_version"] = cluster.SparkVersion
		clusterMap["node_type_id"] = cluster.NodeTypeId
		clusterMap["num_workers"] = cluster.NumWorkers

		result = append(result, clusterMap)
	}

	return result, nil
}

// ListWorkspaces retrieves a list of all workspaces
func (c *Client) ListWorkspaces(ctx context.Context) ([]map[string]interface{}, error) {
	// This is a placeholder as the SDK doesn't directly expose workspace listing
	// In a real implementation, you might use the accounts API or another approach
	log.Println("ListWorkspaces: This is a placeholder method")

	// Return an empty list for now
	return []map[string]interface{}{}, nil
}

// ListJobs retrieves a list of all jobs
func (c *Client) ListJobs(ctx context.Context) ([]map[string]interface{}, error) {
	jobsList, err := c.workspaceClient.Jobs.ListAll(ctx, jobs.ListJobsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	// Convert to map for compatibility with existing code
	result := make([]map[string]interface{}, 0, len(jobsList))
	for _, job := range jobsList {
		jobMap := make(map[string]interface{})
		jobMap["job_id"] = job.JobId
		jobMap["creator_user_name"] = job.CreatorUserName
		jobMap["settings"] = map[string]interface{}{
			"name": job.Settings.Name,
		}

		result = append(result, jobMap)
	}

	return result, nil
}
