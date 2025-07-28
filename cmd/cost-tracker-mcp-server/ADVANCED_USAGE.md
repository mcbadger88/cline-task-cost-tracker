# Advanced Usage - Cost Tracker MCP Server

This document contains detailed information for advanced users, developers, and troubleshooting.

## Alternative Installation Methods

### Development Installation (from source)

```bash
git clone https://github.com/mcbadger88/cline-task-cost-tracker.git
cd cline-task-cost-tracker/cmd/cost-tracker-mcp-server
go mod tidy
go build -o cost-tracker-mcp-server
```

### Running from Source (Development)

For development or if you want to run from source:

```json
{
  "servers": {
    "cost-tracker": {
      "command": "go",
      "args": ["run", "."],
      "env": {},
      "cwd": "/path/to/cline-task-cost-tracker/cmd/cost-tracker-mcp-server",
      "timeout": 300
    }
  }
}
```

## Alternative Configuration Methods

### Project-Level Configuration

Add to your project's Cline MCP settings instead of user-level:

```json
{
  "mcpServers": {
    "cost-tracker": {
      "command": "cost-tracker-mcp-server",
      "args": [],
      "env": {},
      "timeout": 300
    }
  }
}
```

### Using Pre-built Binary with Custom Path

If you have the binary in a custom location:

```json
{
  "servers": {
    "cost-tracker": {
      "command": "/custom/path/to/cost-tracker-mcp-server",
      "args": [],
      "env": {},
      "timeout": 300
    }
  }
}
```

## MCP Tool Reference

### `generate_csv`
Generate CSV file with cost tracking data from ui_messages.json file.

**Parameters:**
- `file_path` (optional): Path to ui_messages.json file. Defaults to current task if not provided.

**Example:**
```json
{
  "name": "generate_csv",
  "arguments": {
    "file_path": "/Users/emma/Library/Application Support/Code/User/globalStorage/saoudrizwan.claude-dev/tasks/1753250474282/ui_messages.json"
  }
}
```

**Returns:** Confirmation message with task ID and CSV generation status.

## CSV Output Format

The server generates CSV files with 15 columns:

1. **Request Summary** - Categorized request types (Task Request, User Input, API Request)
2. **Ask/Say** - Message type and category  
3. **Cost** - Individual API request cost
4. **Text** - Full message content
5. **Timestamp** - Formatted timestamp
6. **Context tokens used** - Token usage information
7. **Total cost** - Cumulative cost
8. **Cline_Action** - Extracted Cline actions
9. **Tool_Used** - Tools invoked during the task
10. **Has_Images** - Whether images were involved
11. **Phase** - Task phase (Initial, Processing, Completion)
12. **Context_Percentage** - Context window usage percentage
13. **Search_Term_In_Transcript** - Unique search identifiers
14. **Cost_Notes** - Additional cost-related notes
15. **Time_Approx** - Approximate time (HH:MM format)

## File Locations

- **CSV Output**: `{repository_root}/ui-log-parser/logs/task_{task_id}_{timestamp}_costs.csv`
- **Monitored Path**: `/Users/emma/Library/Application Support/Code/User/globalStorage/saoudrizwan.claude-dev/tasks/*/ui_messages.json`

## Automatic Repository Detection

The server automatically detects the repository where Cline is currently working and creates CSV logs in the `ui-log-parser/logs` directory within that repository.

### Detection Process

1. **Current Directory Check**: First checks if the MCP server is running within a repository
2. **Repository Search**: Searches common project locations (`~/Projects`, `~/Documents`, `~/Desktop`)
3. **Recent Activity**: Selects the most recently modified repository if multiple are found
4. **Fallback**: Falls back to the MCP server directory if detection fails

### Repository Markers

The server identifies repositories by looking for:
- `.git` directory
- `go.mod` file
- `package.json` file
- `.gitignore` file

### Detection Algorithm

```
1. Check current working directory and walk up tree looking for repo markers
2. If not found, search ~/Projects for repositories
3. If multiple found, select most recently modified
4. If none found, search ~/Documents and ~/Desktop
5. If still none found, fall back to server directory
```

## Validation and Testing

### Compare with CLI Tool

Compare the MCP server's CSV output with the existing CLI tool output to ensure identical results:

```bash
# CLI tool
cd ../cost-tracker-clinerule
go run main.go "/path/to/ui_messages.json"

# MCP server (automatic via file watching or manual via tool)
# Files should be identical
```

### Manual Testing

You can test the server manually:

```bash
# Run the server directly
cost-tracker-mcp-server

# Or from source
cd cmd/cost-tracker-mcp-server
go run .
```

## Troubleshooting

### File Permission Issues
- Ensure the server has read access to the Cline tasks directory
- Ensure write access to the target repository's `ui-log-parser/logs/` directory
- Check directory permissions: `ls -la ~/Library/Application\ Support/Code/User/globalStorage/saoudrizwan.claude-dev/tasks/`

### File Watching Not Working
- Check that the Cline tasks directory exists: `ls ~/Library/Application\ Support/Code/User/globalStorage/saoudrizwan.claude-dev/tasks/`
- Verify task subdirectories contain `ui_messages.json` files
- Look for fsnotify errors in the server logs

### CSV Generation Errors
- Verify that ui_messages.json files are valid JSON: `jq . < ui_messages.json`
- Check that files contain the expected message structure
- Look for parser errors in the server logs

### Repository Detection Issues

If the server can't detect your repository:

**Check Repository Markers**
```bash
# Ensure your project has repository markers
ls -la .git go.mod package.json .gitignore
```

**Debug Repository Detection**
The server logs its detection process:
```
DEBUG: Detected repository root: /path/to/repo
DEBUG: Using ProcessUILogToCSVAutoAt with basePath: /path/to/repo/ui-log-parser
CSV saved to: /path/to/repo/ui-log-parser/logs/
```

**Manual Override**
If detection fails, the server falls back to creating logs in its own directory.

### Debug Logging

Enable verbose logging by checking VSCode's output panel:
1. Open VSCode
2. Go to View → Output
3. Select "Cline" from the dropdown
4. Look for Cost Tracker MCP Server messages

Common debug messages:
- `Cost Tracker MCP Server v2.2.0-auto-detect-repo starting...`
- `Started watching Cline tasks directory`
- `Processing file change: /path/to/ui_messages.json`
- `Successfully processed and generated CSV`

## Architecture

### File Structure
```
cmd/cost-tracker-mcp-server/
├── main.go           # Entry point and signal handling
├── server.go         # MCP protocol implementation
├── handlers.go       # MCP tools implementation
├── file_watcher.go   # File monitoring, debouncing, and repo detection
├── config.go         # Configuration constants
├── README.md         # Quick setup guide
└── ADVANCED_USAGE.md # This file
```

### Component Responsibilities

- **main.go**: Application lifecycle, signal handling, graceful shutdown
- **server.go**: MCP protocol handling, message parsing, tool routing
- **handlers.go**: Implementation of the MCP tool
- **file_watcher.go**: File system monitoring, repository detection, CSV generation
- **config.go**: Constants and configuration values

### Data Flow

```
1. File watcher detects ui_messages.json change
2. Debouncing delays processing by 1 second
3. Repository detection finds target directory
4. ui-log-parser processes file and generates CSV
5. CSV saved to {repo}/ui-log-parser/logs/
```

## Development

### Prerequisites
- Go 1.24.5 or later
- Access to Cline task files

### Running in Development
```bash
cd cmd/cost-tracker-mcp-server
go run .
```

### Building for Distribution
```bash
go build -o cost-tracker-mcp-server
```

### Testing Repository Detection
1. Run the server manually
2. Check debug output for detected repository paths
3. Verify CSV files are created in expected location
4. Test with different repository structures

## Dependencies

- **github.com/fsnotify/fsnotify** - File system watching
- **github.com/mcbadger88/cline-task-cost-tracker/internal/ui-log-parser** - CSV generation logic

## Version History

- **v2.2.0-auto-detect-repo**: Added automatic repository detection and per-repo log organization
- **v2.1.0-eof-fix**: Fixed EOF handling for continuous operation after MCP timeouts
- **v2.0.2-debug-added**: Added comprehensive debug logging
- **v2.0.1-timeout-configured**: Added timeout configuration support
- **v2.0.0**: Initial MCP server implementation with file watching

## Features

- **Automatic File Monitoring**: Watches Cline task directories for `ui_messages.json` changes
- **Debounced Processing**: 1-second delay to avoid processing rapid file changes
- **CSV Generation**: Uses the proven `ui-log-parser` library to generate 15-column CSV files
- **MCP Tool**: Provides a single tool for CSV generation
- **Graceful Error Handling**: Robust error handling for file access issues
- **Repository Auto-Detection**: Automatically finds and organizes logs per repository
- **Continuous Operation**: Survives MCP connection timeouts and keeps running
- **Cross-Repository Support**: Works across all repositories automatically
