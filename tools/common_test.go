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

// NewTestResultError creates a new test result with error content
func NewTestResultError(message string) *TestResult {
	return &TestResult{
		Type:  "error",
		Error: message,
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

// ExtractTestStringParam extracts a string parameter from the test request arguments
func ExtractTestStringParam(request TestRequest, paramName string, defaultValue string) string {
	if val, ok := request.Arguments[paramName].(string); ok {
		return val
	}
	return defaultValue
}

// ExtractTestBoolParam extracts a boolean parameter from the test request arguments
func ExtractTestBoolParam(request TestRequest, paramName string, defaultValue bool) bool {
	if val, ok := request.Arguments[paramName].(bool); ok {
		return val
	}
	return defaultValue
}

// ExtractTestIntParam extracts an integer parameter from the test request arguments
func ExtractTestIntParam(request TestRequest, paramName string, defaultValue int) int {
	if val, ok := request.Arguments[paramName].(float64); ok {
		return int(val)
	}
	return defaultValue
}

// ExtractTestFloatParam extracts a float parameter from the test request arguments
func ExtractTestFloatParam(request TestRequest, paramName string, defaultValue float64) float64 {
	if val, ok := request.Arguments[paramName].(float64); ok {
		return val
	}
	return defaultValue
}

// TestExtractParams tests the parameter extraction functions
func TestExtractParams(t *testing.T) {
	// Create a test request with various parameter types
	request := TestRequest{
		Arguments: map[string]interface{}{
			"string_param":  "test",
			"bool_param":    true,
			"int_param":     float64(42), // JSON numbers are parsed as float64
			"float_param":   float64(3.14),
			"missing_param": nil,
		},
	}

	// Test ExtractTestStringParam
	t.Run("ExtractTestStringParam", func(t *testing.T) {
		// Existing parameter
		result := ExtractTestStringParam(request, "string_param", "default")
		assert.Equal(t, "test", result)

		// Missing parameter
		result = ExtractTestStringParam(request, "nonexistent", "default")
		assert.Equal(t, "default", result)
	})

	// Test ExtractTestBoolParam
	t.Run("ExtractTestBoolParam", func(t *testing.T) {
		// Existing parameter
		result := ExtractTestBoolParam(request, "bool_param", false)
		assert.Equal(t, true, result)

		// Missing parameter
		result = ExtractTestBoolParam(request, "nonexistent", true)
		assert.Equal(t, true, result)
	})

	// Test ExtractTestIntParam
	t.Run("ExtractTestIntParam", func(t *testing.T) {
		// Existing parameter
		result := ExtractTestIntParam(request, "int_param", 0)
		assert.Equal(t, 42, result)

		// Missing parameter
		result = ExtractTestIntParam(request, "nonexistent", 99)
		assert.Equal(t, 99, result)
	})

	// Test ExtractTestFloatParam
	t.Run("ExtractTestFloatParam", func(t *testing.T) {
		// Existing parameter
		result := ExtractTestFloatParam(request, "float_param", 0.0)
		assert.Equal(t, 3.14, result)

		// Missing parameter
		result = ExtractTestFloatParam(request, "nonexistent", 2.71)
		assert.Equal(t, 2.71, result)
	})
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
