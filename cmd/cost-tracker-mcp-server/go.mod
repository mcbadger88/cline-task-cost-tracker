module github.com/mcbadger88/cline-task-cost-tracker/cmd/cost-tracker-mcp-server

go 1.24.5

require (
	github.com/fsnotify/fsnotify v1.7.0
	github.com/mcbadger88/cline-task-cost-tracker/pkg/ui-log-parser v0.0.0
)

require golang.org/x/sys v0.4.0 // indirect

replace github.com/mcbadger88/cline-task-cost-tracker/pkg/ui-log-parser => ../../pkg/ui-log-parser
