# DatabricksMCP Project Guidelines

This document provides guidelines for developing and maintaining the DatabricksMCP project.

## Project Overview

DatabricksMCP is a Go-based project for Databricks Model Context Protocol implementation. The project uses Go 1.24 and provides an interface for interacting with Databricks resources using the MCP pattern.

## Build/Configuration Instructions

### Prerequisites

- Go 1.24 or later
- Git

### Setup

1. Clone the repository:
   ```
   git clone <repository-url>
   cd DatabricksMCP
   ```

2. Install dependencies (if any):
   ```
   go mod tidy
   ```

3. Build the project:
   ```
   go build ./...
   ```

## Testing Information

### Running Tests

To run all tests in the project:

```
go test ./...
```

To run tests with verbose output:

```
go test -v ./...
```

To run tests in a specific package:

```
go test -v ./utils
```

To run a specific test:

```
go test -v ./utils -run TestAdd
```

### Adding New Tests

1. Create test files with the naming convention `*_test.go` in the same package as the code being tested.
2. Use the standard Go testing package `testing`.
3. Follow the table-driven testing pattern for testing multiple cases:

```
// Example test function
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string // replace with appropriate type
        expected string // replace with appropriate type
    }{
        {"test case 1", "input1", "expected1"},
        {"test case 2", "input2", "expected2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Function(tt.input)
            if result != tt.expected {
                t.Errorf("Function(%v) = %v; want %v", tt.input, result, tt.expected)
            }
        })
    }
}
```

### Test Example

The project includes a simple example in the `utils` package:

- `utils.go` contains basic utility functions
- `utils_test.go` demonstrates how to write tests for these functions

To run this example:

```
cd utils
go test -v
```

## Code Style and Development Guidelines

### Code Organization

- Organize code into logical packages
- Keep package names short and descriptive
- Place related files in the same package

### Coding Conventions

- Follow standard Go coding conventions:
  - Use `gofmt` or `go fmt` to format code
  - Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
  - Use meaningful variable and function names
  - Write clear comments for exported functions and types

### Error Handling

- Return errors rather than using panic
- Check all errors and handle them appropriately
- Use custom error types for specific error conditions

### Documentation

- Document all exported functions, types, and constants
- Include examples in documentation where appropriate
- Keep documentation up-to-date with code changes

### Version Control

- Write clear, concise commit messages
- Keep commits focused on a single change
- Use feature branches for new development

## Debugging

- Use the standard Go debugger (`delve`)
- Add logging at appropriate levels
- Consider using environment variables for configuration
