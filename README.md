# Cline Task Cost Tracker

A Go-based cost tracking system for Cline AI assistant tasks, providing detailed CSV reports of API usage and costs.

## Project Structure

This repository follows idiomatic Go project layout with separate packages and commands:

```
├── pkg/
│   └── ui-log-parser/          # Shared library for parsing UI logs
│       ├── types.go            # Data structures
│       ├── parser.go           # Core parsing logic
│       └── csv_writer.go       # CSV output functionality
├── cmd/
│   └── cost-tracker-clinerule/ # CLI tool with .cline_rules integration
│       ├── main.go             # CLI application
│       ├── .cline_rules        # Cline automation rules
│       └── logs/               # Generated CSV files
├── go.mod                      # Go module definition
├── .gitignore                  # Git ignore rules
└── README.md                   # This file
```

## Features

- **Automatic Cost Tracking**: Parses Cline UI message logs to extract API costs
- **Detailed CSV Reports**: Generates comprehensive cost breakdowns with 15 columns
- **Task Identification**: Automatically extracts task IDs and timestamps
- **Large File Handling**: Processes large log files with memory-efficient chunking
- **Reusable Library**: Core logic available as importable Go package
- **Simple API**: Single function call for complete processing

## CSV Output Columns

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

## Usage

### CLI Tool

```bash
cd cmd/cost-tracker-clinerule
go run main.go "/path/to/ui_messages.json"
```

### As a Library

**Simple one-call API (recommended):**
```go
import "github.com/mcbadger88/cline-task-cost-tracker/pkg/ui-log-parser"

// Automatically generate output path and process everything
err := uilogparser.ProcessUILogToCSVAuto(inputPath)

// Or specify custom output path
err := uilogparser.ProcessUILogToCSV(inputPath, outputPath)
```

**Advanced usage (if you need the intermediate data):**
```go
import "github.com/mcbadger88/cline-task-cost-tracker/pkg/ui-log-parser"

// Parse UI messages
messages, err := uilogparser.ParseUIMessages(filePath)
if err != nil {
    log.Fatal(err)
}

// Process into cost records
records := uilogparser.ProcessMessages(messages)

// Generate CSV output
err = uilogparser.WriteCSV(outputPath, records)
```

## Cline Integration

The CLI tool includes `.cline_rules` for automatic cost tracking:

- Automatically executes after every API response
- Works in both Plan Mode and Act Mode
- Generates CSV files in the `logs/` directory
- Uses task start time for consistent filename throughout task duration

## Building

```bash
# Build CLI tool
cd cmd/cost-tracker-clinerule
go build -o cost-tracker

# Run tests
go test ./pkg/ui-log-parser/...

# Tidy dependencies
go mod tidy
```

## Requirements

- Go 1.19 or later
- Access to Cline UI message logs (typically in `~/.../globalStorage/saoudrizwan.claude-dev/tasks/`)

## License

This project is open source and available under standard Go project licensing terms.
