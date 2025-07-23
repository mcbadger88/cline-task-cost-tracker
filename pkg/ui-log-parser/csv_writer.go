package uilogparser

import (
	"encoding/csv"
	"os"
)

// WriteCSV writes cost records to a CSV file
func WriteCSV(filename string, records []CostRecord) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
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

// EnsureLogsDirectory creates the logs directory if it doesn't exist
func EnsureLogsDirectory() error {
	return os.MkdirAll("logs", 0755)
}
