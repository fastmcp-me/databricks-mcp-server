# Databricks MCP Server

A Model Control Protocol (MCP) server for interacting with Databricks.

## Features

- List catalogs in a Databricks workspace
- List schemas in a catalog
- List tables in a schema
- Execute SQL statements

## Installation

You can download the latest release for your platform from the [Releases](https://github.com/yourusername/DatabricksMCP/releases) page.

### Supported Platforms

- Linux (amd64)
- Windows (amd64)
- macOS (Intel/amd64)
- macOS (Apple Silicon/arm64)

## Usage

The application requires Databricks authentication credentials to be set up. You can configure these using environment variables:

```bash
export DATABRICKS_HOST=https://your-workspace.cloud.databricks.com
export DATABRICKS_TOKEN=your-personal-access-token
```

Then run the application:

```bash
./DatabricksMCP
```

## Development

### Prerequisites

- Go 1.24 or later

### Building from Source

```bash
go build
```

### Versioning

This project uses semantic versioning. The version is defined in `version.go` and is automatically updated during the build process based on git tags.

To create a new release:

1. Update the version in `version.go`
2. Commit your changes
3. Create a new tag with the version number:
   ```bash
   git tag v1.2.0
   ```
4. Push the tag to GitHub:
   ```bash
   git push origin v1.2.0
   ```

This will trigger the GitHub Actions workflow to build the binaries for all platforms and create a new release.

## CI/CD Pipeline

This project uses GitHub Actions for continuous integration and delivery. The workflow is defined in `.github/workflows/release.yml` and includes:

- Building the application for multiple platforms
- Running tests
- Creating GitHub releases when a new tag is pushed

## License

[MIT](LICENSE)