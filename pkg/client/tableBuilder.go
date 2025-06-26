package client

import (
	"encoding/json"
	"fmt"
	"strings"
)

// JSONToTable converts JSON array to formatted table
func JSONToTable(jsonStr string) (string, error) {

	// Parse generic JSON
	var data []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("invalid JSON: %v", err)
	}

	if len(data) == 0 {
		return "No data available", nil
	}

	// Get all unique keys (column headers)
	headers := getHeaders(data)
	widths := calculateWidths(data, headers)

	// Build the table
	var builder strings.Builder

	// Table header
	writeRow(&builder, headers, widths)
	builder.WriteString("\n")
	writeSeparator(&builder, widths)
	builder.WriteString("\n")

	// Table rows
	for _, row := range data {
		values := make([]string, len(headers))
		for i, h := range headers {
			if val, exists := row[h]; exists {
				values[i] = fmt.Sprintf("%v", val)
			}
		}
		writeRow(&builder, values, widths)
		builder.WriteString("\n")
	}

	return builder.String(), nil
}

// Helper functions
func getHeaders(data []map[string]interface{}) []string {
	headerMap := make(map[string]bool)
	for _, row := range data {
		for key := range row {
			headerMap[key] = true
		}
	}

	headers := make([]string, 0, len(headerMap))
	for key := range headerMap {
		headers = append(headers, key)
	}
	return headers
}

func calculateWidths(data []map[string]interface{}, headers []string) []int {
	widths := make([]int, len(headers))

	// Initialize with header lengths
	for i, h := range headers {
		widths[i] = len(h)
	}

	// Check data values
	for _, row := range data {
		for i, h := range headers {
			if val, exists := row[h]; exists {
				length := len(fmt.Sprintf("%v", val))
				if length > widths[i] {
					widths[i] = length
				}
			}
		}
	}

	// Add padding
	for i := range widths {
		widths[i] += 2
	}

	return widths
}

func writeRow(b *strings.Builder, values []string, widths []int) {
	b.WriteString("|")
	for i, val := range values {
		fmt.Fprintf(b, " %-*s |", widths[i]-1, val)
	}
}

func writeSeparator(b *strings.Builder, widths []int) {
	b.WriteString("+")
	for _, w := range widths {
		b.WriteString(strings.Repeat("-", w))
		b.WriteString("+")
	}
}
