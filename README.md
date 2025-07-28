# Cost Tracker for Cline

Automatically track Cline task costs across all your repositories with zero configuration.

## What It Does

The Cost Tracker runs in the background and automatically:
- 📊 **Monitors all Cline tasks** across every repository
- 💰 **Generates detailed cost reports** in CSV format
- 📁 **Organizes logs per repository** in `ui-log-parser/logs/`
- 🔄 **Works continuously** without manual intervention

## Quick Setup (Recommended)

### 1. Install the MCP Server

```bash
go install github.com/mcbadger88/cline-task-cost-tracker/cmd/cost-tracker-mcp-server@latest
```

Check: Ensure you have Go's bin directory in your PATH. 

If not:

**macOS/Linux:**
```bash
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc
```


### 2. Add to User MCP Configuration

Add this to `~/Library/Application Support/Code/User/mcp.json`:

```json
{
  "servers": {
    "cost-tracker": {
      "command": "cost-tracker-mcp-server",
      "args": [],
      "env": {},
      "timeout": 300
    }
  }
}
```

### 3. Restart VSCode

After adding the configuration, restart VSCode completely.

### 4. That's It!

The server now runs automatically in the background for all your repositories. Cost tracking CSV files will appear in `ui-log-parser/logs/` within each repository you work on.

## What You Get

### Automatic CSV Reports
Every Cline task generates a detailed CSV file with:
- Individual API request costs
- Cumulative spending
- Token usage statistics
- Tool usage tracking
- Task timing information

### Example Output Location
```
your-project/
├── ui-log-parser/
│   └── logs/
│       └── task_1753250474282_2025-07-23_16-01-14_costs.csv
└── ... (your project files)
```

### MCP Tool Available
Once configured, you'll have access to this tool in Cline:
- `generate_csv` - Generate CSV file with cost tracking data from ui_messages.json file

## Alternative: Cline Rule Installation

If you prefer to use the Cline rule approach instead of the MCP server:

### 1. Copy the Rule File

Copy `cmd/cost-tracker-clinerule/.cline_rules` to your project root.

### 2. Install the CLI Tool

```bash
go install github.com/mcbadger88/cline-task-cost-tracker/cmd/cost-tracker-clinerule@latest
```

### 3. Update the Rule Path

Edit `.cline_rules` to point to your built binary location.

**Note**: The Cline rule approach requires manual setup per project, while the MCP server works globally across all repositories.

## Troubleshooting

**Server not starting?**
- Ensure Go is installed and available in your PATH
- Check that `cost-tracker-mcp-server` is installed: `which cost-tracker-mcp-server`
- Restart VSCode after configuration changes

**No CSV files appearing?**
- The server automatically detects your repository and creates logs there
- Check for `ui-log-parser/logs/` directory in your project root
- Look for debug messages in VSCode's output panel

**Installation issues?**
- Make sure your `GOPATH/bin` is in your system PATH
- If `go install` fails, ensure you have Go 1.21+ installed
- Try running `go clean -modcache` and retry the installation

**Need more details?**
See [cmd/cost-tracker-mcp-server/ADVANCED_USAGE.md](cmd/cost-tracker-mcp-server/ADVANCED_USAGE.md) for detailed configuration options, manual installation, troubleshooting, and development information.

---

**That's it!** Set it once, forget it, and get automatic cost tracking across all your Cline projects.
