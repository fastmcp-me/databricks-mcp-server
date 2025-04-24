# Multi-stage build for Databricks MCP Server

# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application with version information
RUN CGO_ENABLED=0 go build -ldflags="-X 'main.BuildDate=$(date -u +%Y-%m-%d)' -X 'main.GitCommit=$(git rev-parse --short HEAD || echo unknown)'" -o databricks-mcp-server

# Runtime stage
FROM alpine:latest

# Install CA certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/databricks-mcp-server /app/

# Set the entrypoint
ENTRYPOINT ["/app/databricks-mcp-server"]

# Document that the server listens on stdin/stdout
LABEL description="Databricks MCP Server - A Model Context Protocol (MCP) server for interacting with Databricks"
LABEL version="0.0.9"
