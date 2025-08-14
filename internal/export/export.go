package export

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type ExportConfig struct {
	OutputFormat string // json, csv, summary
	OutputFile   string
	TimeFrom     *time.Time
	TimeTo       *time.Time
	ServiceName  string
}

type ExportData struct {
	Traces  []map[string]interface{} `json:"traces,omitempty"`
	Metrics []map[string]interface{} `json:"metrics,omitempty"`
	Logs    []map[string]interface{} `json:"logs,omitempty"`
}

func Export(config ExportConfig) error {
	data := ExportData{}

	// Read traces
	if traces, err := readJSONLFile("./otel-traces.json"); err == nil {
		data.Traces = traces
		fmt.Printf("📊 Found %d trace entries\n", len(traces))
	} else {
		fmt.Printf("⚠️  No traces found: %v\n", err)
	}

	// Read metrics
	if metrics, err := readJSONLFile("./otel-metrics.json"); err == nil {
		data.Metrics = metrics
		fmt.Printf("📈 Found %d metric entries\n", len(metrics))
	} else {
		fmt.Printf("⚠️  No metrics found: %v\n", err)
	}

	// Read logs
	if logs, err := readJSONLFile("./otel-logs.json"); err == nil {
		data.Logs = logs
		fmt.Printf("📝 Found %d log entries\n", len(logs))
	} else {
		fmt.Printf("⚠️  No logs found: %v\n", err)
	}

	if len(data.Traces) == 0 && len(data.Metrics) == 0 && len(data.Logs) == 0 {
		return fmt.Errorf("❌ No telemetry data found. Run 'otel-sandbox up' and 'otel-sandbox verify' first")
	}

	// Filter by service if specified
	if config.ServiceName != "" {
		data = filterByService(data, config.ServiceName)
	}

	switch config.OutputFormat {
	case "summary":
		return exportSummary(data, config.OutputFile)
	case "csv":
		return exportCSV(data, config.OutputFile)
	default:
		return exportJSON(data, config.OutputFile)
	}
}

func readJSONLFile(path string) ([]map[string]interface{}, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data []map[string]interface{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var item map[string]interface{}
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue // Skip malformed entries
		}
		data = append(data, item)
	}

	return data, scanner.Err()
}

func filterByService(data ExportData, serviceName string) ExportData {
	filtered := ExportData{}

	// Filter traces
	for _, trace := range data.Traces {
		if resource, ok := trace["resource"].(map[string]interface{}); ok {
			if attrs, ok := resource["attributes"].(map[string]interface{}); ok {
				if svc, ok := attrs["service.name"]; ok && svc == serviceName {
					filtered.Traces = append(filtered.Traces, trace)
				}
			}
		}
	}

	// Filter metrics
	for _, metric := range data.Metrics {
		if resource, ok := metric["resource"].(map[string]interface{}); ok {
			if attrs, ok := resource["attributes"].(map[string]interface{}); ok {
				if svc, ok := attrs["service.name"]; ok && svc == serviceName {
					filtered.Metrics = append(filtered.Metrics, metric)
				}
			}
		}
	}

	// Filter logs
	for _, log := range data.Logs {
		if resource, ok := log["resource"].(map[string]interface{}); ok {
			if attrs, ok := resource["attributes"].(map[string]interface{}); ok {
				if svc, ok := attrs["service.name"]; ok && svc == serviceName {
					filtered.Logs = append(filtered.Logs, log)
				}
			}
		}
	}

	return filtered
}

func exportJSON(data ExportData, outputFile string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if outputFile == "" {
		fmt.Println(string(jsonData))
		return nil
	}

	fmt.Printf("💾 Exporting to %s\n", outputFile)
	return os.WriteFile(outputFile, jsonData, 0644)
}

func exportSummary(data ExportData, outputFile string) error {
	summary := fmt.Sprintf(`📊 OTel Sandbox Export Summary
=============================
🔍 Traces: %d entries
📈 Metrics: %d entries  
📝 Logs: %d entries

Generated at: %s

Recent Activity:
`, len(data.Traces), len(data.Metrics), len(data.Logs), time.Now().Format(time.RFC3339))

	// Add some sample data if available
	if len(data.Traces) > 0 {
		summary += "\n🔍 Sample Trace Operations:\n"
		seen := make(map[string]bool)
		count := 0
		for _, trace := range data.Traces {
			if spans, ok := trace["spans"].([]interface{}); ok {
				for _, spanInt := range spans {
					if span, ok := spanInt.(map[string]interface{}); ok {
						if name, ok := span["name"].(string); ok && !seen[name] && count < 5 {
							summary += fmt.Sprintf("  - %s\n", name)
							seen[name] = true
							count++
						}
					}
				}
			}
		}
	}

	if outputFile == "" {
		fmt.Println(summary)
		return nil
	}

	fmt.Printf("💾 Exporting summary to %s\n", outputFile)
	return os.WriteFile(outputFile, []byte(summary), 0644)
}

func exportCSV(data ExportData, outputFile string) error {
	var csvData strings.Builder

	// Simple CSV with basic trace info
	csvData.WriteString("Type,Service,Name,Duration,Timestamp\n")

	for _, trace := range data.Traces {
		if spans, ok := trace["spans"].([]interface{}); ok {
			for _, spanInt := range spans {
				if span, ok := spanInt.(map[string]interface{}); ok {
					name := getStringValue(span, "name")
					serviceName := "unknown"

					// Extract service name from resource
					if resource, ok := trace["resource"].(map[string]interface{}); ok {
						if attrs, ok := resource["attributes"].(map[string]interface{}); ok {
							if svc := getStringValue(attrs, "service.name"); svc != "" {
								serviceName = svc
							}
						}
					}

					startTime := getStringValue(span, "startTimeUnixNano")
					endTime := getStringValue(span, "endTimeUnixNano")

					csvData.WriteString(fmt.Sprintf("trace,%s,%s,%s-%s,%s\n",
						serviceName, name, startTime, endTime, startTime))
				}
			}
		}
	}

	if outputFile == "" {
		fmt.Println(csvData.String())
		return nil
	}

	fmt.Printf("💾 Exporting CSV to %s\n", outputFile)
	return os.WriteFile(outputFile, []byte(csvData.String()), 0644)
}

func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", val)
	}
	return ""
}
