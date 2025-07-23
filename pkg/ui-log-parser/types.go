package uilogparser

// UIMessage represents a message from the UI messages log
type UIMessage struct {
	Type      string `json:"type"`
	Say       string `json:"say,omitempty"`
	Ask       string `json:"ask,omitempty"`
	Text      string `json:"text"`
	Timestamp int64  `json:"ts"`
}

// CostRecord represents a row in the cost tracking CSV
type CostRecord struct {
	RequestSummary         string
	AskSay                 string
	Cost                   string
	Text                   string
	Timestamp              string
	ContextTokens          string
	TotalCost              string
	ClineAction            string
	ToolUsed               string
	HasImages              string
	Phase                  string
	ContextPercentage      string
	SearchTermInTranscript string
	CostNotes              string
	TimeApprox             string
}
