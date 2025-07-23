package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

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

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <path_to_ui_messages.json>")
	}

	inputPath := os.Args[1]

	// Check file size before processing
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		log.Fatalf("Error checking file: %v", err)
	}

	fileSizeKB := fileInfo.Size() / 1024
	fmt.Printf("Processing file: %s (Size: %d KB)\n", inputPath, fileSizeKB)

	if fileSizeKB > 500 {
		log.Printf("Warning: Large file detected (%d KB). Processing in chunks to avoid memory issues.", fileSizeKB)
	}

	// Read and parse the JSON file
	data, err := ioutil.ReadFile(inputPath)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	var messages []UIMessage
	if err := json.Unmarshal(data, &messages); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	fmt.Printf("Parsed %d messages\n", len(messages))

	if len(messages) == 0 {
		log.Fatal("No messages found in the file")
	}

	// Process messages and generate CSV records
	records := processMessages(messages)

	// Generate output filename using task start time (first message timestamp)
	taskID := extractTaskID(inputPath)
	startTimestamp := formatTimestampForFilename(messages[0].Timestamp)
	outputPath := fmt.Sprintf("logs/task_%s_%s_costs.csv", taskID, startTimestamp)

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatalf("Error creating logs directory: %v", err)
	}

	// Write CSV file
	if err := writeCSV(outputPath, records); err != nil {
		log.Fatalf("Error writing CSV: %v", err)
	}

	fmt.Printf("Cost tracker CSV generated: %s\n", outputPath)
	fmt.Printf("Total records: %d\n", len(records))
}

func processMessages(messages []UIMessage) []CostRecord {
	var records []CostRecord
	var totalCost float64

	for i, msg := range messages {
		record := CostRecord{
			Timestamp: formatTimestamp(msg.Timestamp),
			Text:      msg.Text,
		}

		// Determine Ask/Say column and Request Summary
		if msg.Type == "say" {
			record.AskSay = fmt.Sprintf(`"say": "%s"`, msg.Say)
			record.RequestSummary = categorizeMessage(msg.Say, msg.Text, i)
		} else if msg.Type == "ask" {
			record.AskSay = fmt.Sprintf(`"ask": "%s"`, msg.Ask)
			record.RequestSummary = "" // Ask messages don't populate Request Summary
		}

		// Extract cost information
		cost := extractCost(msg.Text)
		if cost > 0 {
			record.Cost = fmt.Sprintf("%.6f", cost)
			totalCost += cost
		}

		// Extract context tokens
		record.ContextTokens = extractContextTokens(msg.Text)

		// Set total cost (cumulative)
		record.TotalCost = fmt.Sprintf("%.6f", totalCost)

		// Generate additional fields
		record.ClineAction = extractClineAction(msg)
		record.ToolUsed = extractToolUsed(msg)
		record.HasImages = extractHasImages(msg)
		record.Phase = determinePhase(msg, i)
		record.ContextPercentage = extractContextPercentage(msg)
		record.SearchTermInTranscript = generateSearchTerm(msg, i)
		record.CostNotes = generateCostNotes(msg)
		record.TimeApprox = formatTimeApprox(msg.Timestamp)

		records = append(records, record)
	}

	return records
}

func categorizeMessage(sayType, text string, index int) string {
	// Only populate Request Summary for specific cases
	switch sayType {
	case "user_feedback":
		return fmt.Sprintf("User Input: %s", text)
	case "api_req_started":
		return fmt.Sprintf("API Request: %s", text)
	default:
		// Check if this is the first row
		if index == 0 {
			return fmt.Sprintf("Task Request: %s", text)
		}
		// Leave blank for all other rows
		return ""
	}
}

func categorizeAskMessage(askType, text string) string {
	switch askType {
	case "tool":
		return "Tool execution"
	case "command":
		return "Command execution"
	case "command_output":
		return "Command output"
	case "followup":
		return "Follow-up question"
	default:
		return "User interaction"
	}
}

func extractCost(text string) float64 {
	// Look for cost patterns in the text
	costPatterns := []string{
		`"cost":\s*([0-9.]+)`,
		`cost":\s*([0-9.]+)`,
		`\$([0-9.]+)`,
		`"totalCost":\s*([0-9.]+)`,
	}

	for _, pattern := range costPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			if cost, err := strconv.ParseFloat(matches[1], 64); err == nil {
				return cost
			}
		}
	}

	return 0
}

func extractContextTokens(text string) string {
	// Look for token usage patterns
	tokenPatterns := []string{
		`"inputTokens":\s*([0-9]+)`,
		`"outputTokens":\s*([0-9]+)`,
		`([0-9]+)\s*tokens`,
		`context.*?([0-9]+).*?tokens`,
	}

	for _, pattern := range tokenPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

func extractClineAction(msg UIMessage) string {
	if msg.Type == "say" && msg.Say == "text" {
		text := msg.Text
		if len(text) > 100 {
			text = text[:100] + "..."
		}
		return strings.ReplaceAll(text, "\"", "'")
	}
	return ""
}

func extractToolUsed(msg UIMessage) string {
	if msg.Type == "ask" && msg.Ask == "tool" {
		// Extract tool name from JSON in text
		re := regexp.MustCompile(`"tool":"([^"]+)"`)
		matches := re.FindStringSubmatch(msg.Text)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

func extractHasImages(msg UIMessage) string {
	if strings.Contains(msg.Text, "images") || strings.Contains(msg.Text, "image") {
		return "Yes"
	}
	return "No"
}

func determinePhase(msg UIMessage, index int) string {
	if index < 5 {
		return "Initial"
	} else if msg.Type == "say" && msg.Say == "api_req_started" {
		return "Processing"
	} else if strings.Contains(msg.Text, "completion") || strings.Contains(msg.Text, "finished") {
		return "Completion"
	}
	return "Processing"
}

func extractContextPercentage(msg UIMessage) string {
	// Look for context usage patterns
	re := regexp.MustCompile(`([0-9]+)%`)
	matches := re.FindStringSubmatch(msg.Text)
	if len(matches) > 1 {
		return matches[1] + "%"
	}
	return ""
}

func generateSearchTerm(msg UIMessage, index int) string {
	// Generate unique search term based on message content
	if msg.Type == "say" && msg.Say == "user_feedback" {
		words := strings.Fields(msg.Text)
		if len(words) > 2 {
			return fmt.Sprintf("user_%s_%s_%d", words[0], words[1], index)
		}
	}
	return fmt.Sprintf("%s_%s_%d", msg.Type, msg.Say, index)
}

func generateCostNotes(msg UIMessage) string {
	if strings.Contains(msg.Text, "cost") {
		return "Has cost data"
	}
	return ""
}

func formatTimestamp(ts int64) string {
	t := time.Unix(ts/1000, (ts%1000)*1000000)
	return t.Format("2006-01-02 15:04:05")
}

func formatTimestampForFilename(ts int64) string {
	t := time.Unix(ts/1000, (ts%1000)*1000000)
	return t.Format("2006-01-02_15-04-05")
}

func formatTimeApprox(ts int64) string {
	t := time.Unix(ts/1000, (ts%1000)*1000000)
	return t.Format("15:04")
}

func extractTaskID(path string) string {
	// Extract task ID from path
	re := regexp.MustCompile(`tasks/([0-9]+)/`)
	matches := re.FindStringSubmatch(path)
	if len(matches) > 1 {
		return matches[1]
	}
	return "unknown"
}

func writeCSV(filename string, records []CostRecord) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header (without User_Request column)
	header := []string{
		"Request Summary", "Ask/Say", "Cost", "Text", "Timestamp",
		"Context tokens used", "Total cost", "Cline_Action",
		"Tool_Used", "Has_Images", "Phase", "Context_Percentage",
		"Search_Term_In_Transcript", "Cost_Notes", "Time_Approx",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write records
	for _, record := range records {
		row := []string{
			record.RequestSummary,
			record.AskSay,
			record.Cost,
			record.Text,
			record.Timestamp,
			record.ContextTokens,
			record.TotalCost,
			record.ClineAction,
			record.ToolUsed,
			record.HasImages,
			record.Phase,
			record.ContextPercentage,
			record.SearchTermInTranscript,
			record.CostNotes,
			record.TimeApprox,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
