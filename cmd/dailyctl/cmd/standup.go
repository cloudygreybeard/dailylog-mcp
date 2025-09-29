package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"dailylog/internal/storage"

	"github.com/spf13/cobra"
)

// standupCmd represents the standup command
var standupCmd = &cobra.Command{
	Use:   "standup",
	Short: "Generate standup reports in various formats",
	Long: `Generate standup reports for daily team meetings.

Supports multiple output formats including Slack-style YAML format.

Examples:
  dailyctl standup --format slack-yaml
  dailyctl standup --format slack-yaml --copy
  dailyctl standup --format json`,
	RunE: runStandupReport,
}

func init() {
	rootCmd.AddCommand(standupCmd)

	standupCmd.Flags().String("format", "default", "Output format: default, slack-yaml, json")
	standupCmd.Flags().Bool("copy", false, "Copy output to clipboard (macOS)")
	standupCmd.Flags().String("date", "", "Date for standup (YYYY-MM-DD, defaults to today)")
}

func runStandupReport(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("format")
	copyToClipboard, _ := cmd.Flags().GetBool("copy")
	dateStr, _ := cmd.Flags().GetString("date")

	// Parse date
	var targetDate time.Time
	if dateStr != "" {
		var err error
		targetDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return fmt.Errorf("invalid date format: %s (use YYYY-MM-DD)", dateStr)
		}
	} else {
		targetDate = time.Now()
	}

	// Create storage provider
	storageProvider, err := createStorageProvider()
	if err != nil {
		return fmt.Errorf("failed to create storage provider: %v", err)
	}

	// Get yesterday's entries (what was done)
	yesterday := targetDate.AddDate(0, 0, -1)
	yesterdayLog, err := storageProvider.GetDay(yesterday)
	if err != nil {
		return fmt.Errorf("failed to get yesterday's entries: %v", err)
	}

	// Get today's entries (what's planned)
	todayLog, err := storageProvider.GetDay(targetDate)
	if err != nil {
		return fmt.Errorf("failed to get today's entries: %v", err)
	}

	// Generate standup report
	report := generateStandupReport(yesterdayLog.Entries, todayLog.Entries, format, targetDate)

	if copyToClipboard {
		// Copy to clipboard (macOS)
		err := copyToClipboardMacOS(report)
		if err != nil {
			fmt.Printf("Warning: Could not copy to clipboard: %v\n\n", err)
		} else {
			fmt.Println("Report copied to clipboard!")
			fmt.Println("")
		}
	}

	fmt.Print(report)
	return nil
}

func generateStandupReport(yesterdayEntries, todayEntries []storage.DailyLogEntry, format string, date time.Time) string {
	switch format {
	case "slack-yaml":
		return generateSlackYAMLReport(yesterdayEntries, todayEntries, date)
	case "json":
		return generateJSONReport(yesterdayEntries, todayEntries, date)
	default:
		return generateDefaultReport(yesterdayEntries, todayEntries, date)
	}
}

func generateSlackYAMLReport(yesterdayEntries, todayEntries []storage.DailyLogEntry, date time.Time) string {
	var report strings.Builder

	yesterday := date.AddDate(0, 0, -1)

	report.WriteString(fmt.Sprintf("Standup Report - %s\n", date.Format("2006-01-02")))
	report.WriteString("```yaml\n")

	// Yesterday's work
	report.WriteString(fmt.Sprintf("Y: # Yesterday (%s)\n", yesterday.Format("Jan 2")))
	if len(yesterdayEntries) == 0 {
		report.WriteString("  - No activities recorded\n")
	} else {
		for _, entry := range yesterdayEntries {
			if entry.Type == "activity" {
				status := ""
				if entry.Status > 0 {
					status = fmt.Sprintf(" (status: %d/10)", entry.Status)
				}
				report.WriteString(fmt.Sprintf("  - %s%s\n", entry.Title, status))
			}
		}
	}

	report.WriteString("\n")

	// Today's plan
	report.WriteString(fmt.Sprintf("T: # Today (%s)\n", date.Format("Jan 2")))
	todayPlanned := filterPlannedEntries(todayEntries)
	if len(todayPlanned) == 0 {
		report.WriteString("  - Planning session\n")
	} else {
		for _, entry := range todayPlanned {
			priority := ""
			if entry.Priority > 0 {
				priority = fmt.Sprintf(" (priority: %d/5)", entry.Priority)
			}
			report.WriteString(fmt.Sprintf("  - %s%s\n", entry.Title, priority))
		}
	}

	report.WriteString("```")

	return report.String()
}

func generateJSONReport(yesterdayEntries, todayEntries []storage.DailyLogEntry, date time.Time) string {
	yesterday := date.AddDate(0, 0, -1)

	report := map[string]interface{}{
		"date": date.Format("2006-01-02"),
		"yesterday": map[string]interface{}{
			"date":       yesterday.Format("2006-01-02"),
			"activities": filterActivities(yesterdayEntries),
		},
		"today": map[string]interface{}{
			"date":    date.Format("2006-01-02"),
			"planned": filterPlannedEntries(todayEntries),
		},
	}

	return formatJSON(report)
}

func generateDefaultReport(yesterdayEntries, todayEntries []storage.DailyLogEntry, date time.Time) string {
	var report strings.Builder
	yesterday := date.AddDate(0, 0, -1)

	report.WriteString(fmt.Sprintf("Standup Report - %s\n", date.Format("2006-01-02")))
	report.WriteString(strings.Repeat("=", 40))
	report.WriteString("\n\n")

	// Yesterday
	report.WriteString(fmt.Sprintf("Yesterday (%s):\n", yesterday.Format("Jan 2")))
	if len(yesterdayEntries) == 0 {
		report.WriteString("  • No activities recorded\n")
	} else {
		for _, entry := range yesterdayEntries {
			if entry.Type == "activity" {
				report.WriteString(fmt.Sprintf("  • %s\n", entry.Title))
			}
		}
	}

	report.WriteString("\n")

	// Today
	report.WriteString(fmt.Sprintf("Today (%s):\n", date.Format("Jan 2")))
	todayPlanned := filterPlannedEntries(todayEntries)
	if len(todayPlanned) == 0 {
		report.WriteString("  • Planning session\n")
	} else {
		for _, entry := range todayPlanned {
			report.WriteString(fmt.Sprintf("  • %s\n", entry.Title))
		}
	}

	return report.String()
}

func filterActivities(entries []storage.DailyLogEntry) []storage.DailyLogEntry {
	var activities []storage.DailyLogEntry
	for _, entry := range entries {
		if entry.Type == "activity" {
			activities = append(activities, entry)
		}
	}
	return activities
}

func filterPlannedEntries(entries []storage.DailyLogEntry) []storage.DailyLogEntry {
	var planned []storage.DailyLogEntry
	for _, entry := range entries {
		// Include activities marked as planned, or if no planned activities, include all activities
		if entry.Type == "activity" {
			planned = append(planned, entry)
		}
	}
	return planned
}

func copyToClipboardMacOS(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func formatJSON(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting JSON: %v", err)
	}
	return string(data)
}
