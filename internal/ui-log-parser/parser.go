package uilogparser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseUIMessages reads and parses a UI messages JSON file
func ParseUIMessages(filePath string) ([]UIMessage, error) {
	// Check file size before processing
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("error checking file: %v", err)
	}

	fileSizeKB := fileInfo.Size() / 1024
	fmt.Printf("Processing file: %s (Size: %d KB)\n", filePath, fileSizeKB)

	if fileSizeKB > 500 {
		log.Printf("Warning: Large file detected (%d KB). Processing in chunks to avoid memory issues.", fileSizeKB)
	}

	// Read and parse the JSON file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	var messages []UIMessage
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	fmt.Printf("Parsed %d messages\n", len(messages))

	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages found in the file")
	}

	return messages, nil
}

// ProcessMessages converts UI messages to cost records
func ProcessMessages(messages []UIMessage) []CostRecord {
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

// ExtractTaskID extracts task ID from file path
func ExtractTaskID(path string) string {
	re := regexp.MustCompile(`tasks/([0-9]+)/`)
	matches := re.FindStringSubmatch(path)
	if len(matches) > 1 {
		return matches[1]
	}
	return "unknown"
}

// GenerateOutputPath creates the output CSV file path
func GenerateOutputPath(taskID string, startTimestamp int64) string {
	startTime := formatTimestampForFilename(startTimestamp)
	return fmt.Sprintf("logs/task_%s_%s_costs.csv", taskID, startTime)
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

// extractWorkingDirectoryFromFile extracts the working directory from ui_messages.json
func extractWorkingDirectoryFromFile(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("DEBUG: Error opening file for working directory extraction: %v", err)
		return ""
	}
	defer file.Close()

	// Read the file content
	content := make([]byte, 10240) // Read first 10KB to find the working directory
	n, err := file.Read(content)
	if err != nil && n == 0 {
		log.Printf("DEBUG: Error reading file for working directory extraction: %v", err)
		return ""
	}

	contentStr := string(content[:n])

	// Look for the pattern "# Current Working Directory (/path/to/directory)"
	workingDirPattern := "# Current Working Directory ("
	startIdx := strings.Index(contentStr, workingDirPattern)
	if startIdx == -1 {
		log.Printf("DEBUG: Working directory pattern not found in file")
		return ""
	}

	// Find the start of the path
	pathStart := startIdx + len(workingDirPattern)

	// Find the end of the path (closing parenthesis)
	pathEnd := strings.Index(contentStr[pathStart:], ")")
	if pathEnd == -1 {
		log.Printf("DEBUG: Could not find end of working directory path in file")
		return ""
	}

	workingDir := contentStr[pathStart : pathStart+pathEnd]
	log.Printf("DEBUG: Extracted working directory from file: %s", workingDir)
	return workingDir
}

// ProcessMessagesWithWorkingDir converts UI messages to cost records with working directory
func ProcessMessagesWithWorkingDir(messages []UIMessage, workingDir string) []CostRecord {
	var records []CostRecord
	var totalCost float64

	for i, msg := range messages {
		record := CostRecord{
			Timestamp:        formatTimestamp(msg.Timestamp),
			Text:             msg.Text,
			WorkingDirectory: workingDir,
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

// ProcessUILogToCSV is a convenience function that handles the entire pipeline
// from UI messages JSON file to CSV output in a single call
func ProcessUILogToCSV(inputPath, outputPath string) error {
	// Parse UI messages
	messages, err := ParseUIMessages(inputPath)
	if err != nil {
		return err
	}

	// Process messages into cost records
	records := ProcessMessages(messages)

	// Ensure logs directory exists
	if err := EnsureLogsDirectory(); err != nil {
		return err
	}

	// Write CSV file
	if err := WriteCSV(outputPath, records); err != nil {
		return err
	}

	fmt.Printf("Cost tracker CSV generated: %s\n", outputPath)
	fmt.Printf("Total records: %d\n", len(records))
	return nil
}

// ProcessUILogToCSVAuto automatically generates the output path based on input
func ProcessUILogToCSVAuto(inputPath string) error {
	// Parse just enough to get the first message for timestamp
	messages, err := ParseUIMessages(inputPath)
	if err != nil {
		return err
	}

	if len(messages) == 0 {
		return fmt.Errorf("no messages found in the file")
	}

	// Generate output path
	taskID := ExtractTaskID(inputPath)
	outputPath := GenerateOutputPath(taskID, messages[0].Timestamp)

	// Process the rest
	records := ProcessMessages(messages)

	// Ensure logs directory exists
	if err := EnsureLogsDirectory(); err != nil {
		return err
	}

	// Write CSV file
	if err := WriteCSV(outputPath, records); err != nil {
		return err
	}

	fmt.Printf("Cost tracker CSV generated: %s\n", outputPath)
	fmt.Printf("Total records: %d\n", len(records))
	return nil
}

// ProcessUILogToCSVAutoAt automatically generates the output path based on input
// and creates the logs directory at the specified base path
func ProcessUILogToCSVAutoAt(inputPath, basePath string) error {
	log.Printf("DEBUG: ProcessUILogToCSVAutoAt called with inputPath=%s, basePath=%s", inputPath, basePath)

	// Parse just enough to get the first message for timestamp
	messages, err := ParseUIMessages(inputPath)
	if err != nil {
		return err
	}

	if len(messages) == 0 {
		return fmt.Errorf("no messages found in the file")
	}

	// Extract working directory from the input file
	workingDir := extractWorkingDirectoryFromFile(inputPath)
	log.Printf("DEBUG: Extracted working directory: %s", workingDir)

	// Generate output path relative to base path
	taskID := ExtractTaskID(inputPath)
	outputFilename := fmt.Sprintf("task_%s_%s_costs.csv", taskID, formatTimestampForFilename(messages[0].Timestamp))
	outputPath := filepath.Join(basePath, "logs", outputFilename)

	log.Printf("DEBUG: Generated outputPath: %s", outputPath)

	// Process the rest with working directory
	records := ProcessMessagesWithWorkingDir(messages, workingDir)

	// Ensure logs directory exists at the specified base path
	log.Printf("DEBUG: About to call EnsureLogsDirectoryAt with basePath: %s", basePath)
	if err := EnsureLogsDirectoryAt(basePath); err != nil {
		log.Printf("DEBUG: EnsureLogsDirectoryAt failed: %v", err)
		return err
	}
	log.Printf("DEBUG: EnsureLogsDirectoryAt succeeded")

	// Write CSV file
	if err := WriteCSV(outputPath, records); err != nil {
		return err
	}

	fmt.Printf("Cost tracker CSV generated: %s\n", outputPath)
	fmt.Printf("Total records: %d\n", len(records))
	return nil
}
