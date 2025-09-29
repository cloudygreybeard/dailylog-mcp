package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// outputJSON outputs data as formatted JSON
func outputJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// outputYAML outputs data as formatted YAML
func outputYAML(data interface{}) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %v", err)
	}
	fmt.Println(string(yamlData))
	return nil
}

// formatDuration formats a duration in minutes to a human-readable string
func formatDuration(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	remainingMinutes := minutes % 60
	if remainingMinutes == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh%dm", hours, remainingMinutes)
}

// truncate truncates a string to the specified length with ellipsis
func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

// formatTags formats a slice of tags as a comma-separated string
func formatTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return strings.Join(tags, ",")
}

// formatMood formats a mood value as a string
func formatMood(mood int) string {
	if mood == 0 {
		return ""
	}
	return fmt.Sprintf("%d/10", mood)
}

// formatPriority formats a priority value as a string
func formatPriority(priority int) string {
	if priority == 0 {
		return ""
	}
	return fmt.Sprintf("%d/5", priority)
}

// padRight pads a string to the right with spaces
func padRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-len(s))
}

// padLeft pads a string to the left with spaces
func padLeft(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat(" ", length-len(s)) + s
}
