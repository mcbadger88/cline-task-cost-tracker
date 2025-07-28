# Cost Tracker MCP Server

Automatically track Cline task costs across all your repositories with zero configuration.

## What It Does

The Cost Tracker MCP Server runs in the background and automatically:
- 📊 **Monitors all Cline tasks** across every repository
- 💰 **Generates detailed cost reports** in CSV format
- 📁 **Organizes logs per repository** in `ui-log-parser/logs/`
- 🔄 **Works continuously** without manual intervention

## Quick Setup

### 1. Install the Server

```bash
go install github.com/mcbadger88/cline-task-cost-tracker/cmd/cost-tracker-mcp-server@latest
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
- Try reinstalling: `go install github.com/mcbadger88/cline-task-cost-tracker/cmd/cost-tracker-mcp-server@latest`

**Need more details?**
See [ADVANCED_USAGE.md](ADVANCED_USAGE.md) for detailed configuration options, manual installation, troubleshooting, and development information.

---

**That's it!** Set it once, forget it, and get automatic cost tracking across all your Cline projects.
