package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRequest is a simplified version of mcp.CallToolRequest for testing
type TestRequest struct {
	Arguments map[string]interface{}
}

// TestResult is a simplified version of mcp.CallToolResult for testing
type TestResult struct {
	Type  string
	Text  string
	Error string
}

// TestOperation is a simplified version of DatabricksTool for testing
type TestOperation func(ctx context.Context, request TestRequest) (interface{}, error)

// NewTestResultText creates a new test result with text content
func NewTestResultText(text string) *TestResult {
	return &TestResult{
		Type: "text",
		Text: text,
	}
}

// NewTestResultErrorFromErr creates a new test result with error content from an error
func NewTestResultErrorFromErr(message string, err error) *TestResult {
	return &TestResult{
		Type:  "error",
		Error: message + ": " + err.Error(),
	}
}

// ExecuteTestOperation is a simplified version of ExecuteTool for testing
func ExecuteTestOperation(operation TestOperation) func(ctx context.Context, request TestRequest) (*TestResult, error) {
	return func(ctx context.Context, request TestRequest) (*TestResult, error) {
		result, err := operation(ctx, request)
		if err != nil {
			return NewTestResultErrorFromErr("Error executing operation", err), nil
		}

		// Marshal the result to JSON
		jsonResult, err := json.Marshal(result)
		if err != nil {
			return NewTestResultErrorFromErr("Error marshalling result into JSON", err), nil
		}

		return NewTestResultText(string(jsonResult)), nil
	}
}

// TestExecuteTestOperation tests the ExecuteTestOperation function
func TestExecuteTestOperation(t *testing.T) {
	// Create a mock operation that returns a successful result
	successOp := func(ctx context.Context, request TestRequest) (interface{}, error) {
		return map[string]string{"result": "success"}, nil
	}

	// Create a mock operation that returns an error
	errorOp := func(ctx context.Context, request TestRequest) (interface{}, error) {
		return nil, errors.New("operation failed")
	}

	// Test successful operation
	t.Run("SuccessfulOperation", func(t *testing.T) {
		handler := ExecuteTestOperation(successOp)
		result, err := handler(context.Background(), TestRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "text", result.Type)
		assert.NotEmpty(t, result.Text)
	})

	// Test failed operation
	t.Run("FailedOperation", func(t *testing.T) {
		handler := ExecuteTestOperation(errorOp)
		result, err := handler(context.Background(), TestRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "error", result.Type)
		assert.Contains(t, result.Error, "operation failed")
	})
}
