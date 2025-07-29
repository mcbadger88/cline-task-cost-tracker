# Cost Tracker MCP Server (SDK Version)

A Model Context Protocol (MCP) server for automatic cost tracking of Cline tasks, built using the official Go MCP SDK.

## Features

- **Official MCP SDK**: Uses `github.com/modelcontextprotocol/go-sdk` for proper MCP protocol compliance
- **Automatic File Watching**: Monitors Cline task files with 1-second debouncing
- **CSV Generation**: Automatically generates cost tracking CSV files
- **Repository Detection**: Intelligently detects working directories and places CSV files appropriately
- **Single Tool**: Provides `generate_csv` tool for manual CSV generation

## Installation

```bash
go install github.com/mcbadger88/cline-task-cost-tracker/cmd/cost-tracker-mcp-server-sdk@latest
```

## Usage

```bash
cost-tracker-mcp-server-sdk
```

## Tools

### generate_csv

Generates a CSV file with cost tracking data from a Cline task's ui_messages.json file.

**Parameters:**
- `file_path` (optional): Path to ui_messages.json file. If not provided, uses the most recent task.

**Example:**
```json
{
  "file_path": "/path/to/task/ui_messages.json"
}
```

## File Watching

The server automatically monitors the Cline tasks directory (`~/.cline/tasks/`) and:

1. Watches for changes to `ui_messages.json` files
2. Uses 1-second debouncing to avoid processing rapid changes
3. Automatically generates CSV files when tasks are updated
4. Places CSV files in `{detected_repo}/ui-log-parser/logs/`

## Architecture

- **main.go**: Entry point with graceful shutdown handling
- **server.go**: MCP server implementation using official SDK
- **tools.go**: Tool implementations (generate_csv)
- **watcher.go**: File watching logic with debouncing
- **config.go**: Configuration constants and utilities

## Dependencies

- `github.com/modelcontextprotocol/go-sdk`: Official MCP SDK
- `github.com/fsnotify/fsnotify`: File system notifications
- `github.com/mcbadger88/cline-task-cost-tracker/pkg/ui-log-parser`: CSV generation logic

## Differences from Manual Implementation

This SDK version provides:
- ✅ Proper MCP protocol compliance
- ✅ Better error handling with SDK patterns
- ✅ Cleaner, more maintainable code
- ✅ Official SDK support and updates
- ✅ Standardized MCP server patterns

## Usage with Cline

Add this server to your MCP configuration to enable automatic cost tracking for all Cline tasks.
