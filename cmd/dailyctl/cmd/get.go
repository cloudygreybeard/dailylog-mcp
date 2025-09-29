package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"dailylog/internal/storage"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get log entries",
	Long: `Get log entries for specific dates or date ranges.

Examples:
  dailyctl get today
  dailyctl get yesterday  
  dailyctl get 2025-09-29
  dailyctl get --date-start 2025-09-01 --date-end 2025-09-30
  dailyctl get week
  dailyctl get month`,
}

var getTodayCmd = &cobra.Command{
	Use:   "today",
	Short: "Get today's entries",
	RunE:  runGetEntries("today"),
}

var getYesterdayCmd = &cobra.Command{
	Use:   "yesterday",
	Short: "Get yesterday's entries",
	RunE:  runGetEntries("yesterday"),
}

var getWeekCmd = &cobra.Command{
	Use:   "week",
	Short: "Get this week's entries",
	RunE:  runGetEntries("week"),
}

var getMonthCmd = &cobra.Command{
	Use:   "month",
	Short: "Get this month's entries",
	RunE:  runGetEntries("month"),
}

var getDateCmd = &cobra.Command{
	Use:   "date [YYYY-MM-DD]",
	Short: "Get entries for a specific date",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		date := args[0]
		return getEntriesForDate(date)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Add subcommands
	getCmd.AddCommand(getTodayCmd)
	getCmd.AddCommand(getYesterdayCmd)
	getCmd.AddCommand(getWeekCmd)
	getCmd.AddCommand(getMonthCmd)
	getCmd.AddCommand(getDateCmd)

	// Add flags
	getCmd.PersistentFlags().String("date-start", "", "Start date for range query (YYYY-MM-DD)")
	getCmd.PersistentFlags().String("date-end", "", "End date for range query (YYYY-MM-DD)")
	getCmd.PersistentFlags().String("type", "", "Filter by entry type")
	getCmd.PersistentFlags().StringSlice("tags", []string{}, "Filter by tags")
	getCmd.PersistentFlags().Int("limit", 0, "Maximum number of entries to return")
	getCmd.PersistentFlags().Bool("stats", false, "Include summary statistics")
}

func runGetEntries(period string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var targetDate time.Time
		var dateStart, dateEnd *time.Time

		now := time.Now()

		switch period {
		case "today":
			targetDate = now
		case "yesterday":
			targetDate = now.AddDate(0, 0, -1)
		case "week":
			// Get week start (Monday)
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7 // Sunday = 7
			}
			weekStart := now.AddDate(0, 0, -(weekday - 1))
			weekEnd := weekStart.AddDate(0, 0, 6)
			dateStart = &weekStart
			dateEnd = &weekEnd
		case "month":
			monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			monthEnd := monthStart.AddDate(0, 1, -1)
			dateStart = &monthStart
			dateEnd = &monthEnd
		}

		return getEntries(cmd, targetDate, dateStart, dateEnd)
	}
}

func getEntriesForDate(dateStr string) error {
	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format: %s (use YYYY-MM-DD)", dateStr)
	}

	return getEntries(nil, targetDate, nil, nil)
}

func getEntries(cmd *cobra.Command, targetDate time.Time, dateStart, dateEnd *time.Time) error {
	// Create storage provider
	storageProvider, err := createStorageProvider()
	if err != nil {
		return fmt.Errorf("failed to create storage provider: %v", err)
	}

	var entries []storage.DailyLogEntry
	var period string

	if dateStart != nil && dateEnd != nil {
		// Get entries for date range
		searchReq := storage.LogSearchRequest{
			DateStart: dateStart,
			DateEnd:   dateEnd,
		}

		if cmd != nil {
			// Apply filters from flags
			entryType, _ := cmd.Flags().GetString("type")
			tags, _ := cmd.Flags().GetStringSlice("tags")
			limit, _ := cmd.Flags().GetInt("limit")

			searchReq.Type = entryType
			searchReq.Tags = tags
			searchReq.Limit = limit
		}

		searchResult, err := storageProvider.SearchLogs(searchReq)
		if err != nil {
			return fmt.Errorf("failed to search logs: %v", err)
		}

		entries = searchResult.Entries
		period = fmt.Sprintf("%s to %s", dateStart.Format("2006-01-02"), dateEnd.Format("2006-01-02"))

	} else {
		// Get entries for specific date
		dayLog, err := storageProvider.GetDay(targetDate)
		if err != nil {
			return fmt.Errorf("failed to get day: %v", err)
		}

		entries = dayLog.Entries
		period = targetDate.Format("2006-01-02")
	}

	// Output results
	outputFormat := viper.GetString("output.format")
	switch outputFormat {
	case "json":
		return outputJSON(map[string]interface{}{
			"entries":     entries,
			"total_count": len(entries),
			"period":      period,
		})
	case "yaml":
		return outputYAML(map[string]interface{}{
			"entries":     entries,
			"total_count": len(entries),
			"period":      period,
		})
	default:
		return outputEntriesTable(entries, period, cmd)
	}
}

func outputEntriesTable(entries []storage.DailyLogEntry, period string, cmd *cobra.Command) error {
	fmt.Printf("Daily Log Entries - %s\n", period)
	fmt.Printf("=====================%s\n", strings.Repeat("=", len(period)))
	fmt.Println()

	if len(entries) == 0 {
		fmt.Println("No entries found.")
		return nil
	}

	// Table header
	fmt.Printf("%-12s %-8s %-15s %-30s %-20s %-6s %-8s\n",
		"TIME", "TYPE", "TAGS", "TITLE", "LOCATION", "MOOD", "PRIORITY")
	fmt.Println(strings.Repeat("-", 100))

	// Table rows
	for _, entry := range entries {
		timeStr := entry.Timestamp.Format("15:04:05")
		tagsStr := strings.Join(entry.Tags, ",")
		if len(tagsStr) > 15 {
			tagsStr = tagsStr[:12] + "..."
		}
		titleStr := entry.Title
		if len(titleStr) > 30 {
			titleStr = titleStr[:27] + "..."
		}
		locationStr := entry.Location
		if len(locationStr) > 20 {
			locationStr = locationStr[:17] + "..."
		}

		moodStr := ""
		if entry.Mood > 0 {
			moodStr = fmt.Sprintf("%d/10", entry.Mood)
		}

		priorityStr := ""
		if entry.Priority > 0 {
			priorityStr = fmt.Sprintf("%d/5", entry.Priority)
		}

		fmt.Printf("%-12s %-8s %-15s %-30s %-20s %-6s %-8s\n",
			timeStr, entry.Type, tagsStr, titleStr, locationStr, moodStr, priorityStr)

		// Show description if available and not too long
		if entry.Description != "" && len(entry.Description) <= 80 {
			fmt.Printf("             %s\n", entry.Description)
		}
	}

	fmt.Println()
	fmt.Printf("Total entries: %d\n", len(entries))

	// Show stats if requested
	if cmd != nil {
		showStats, _ := cmd.Flags().GetBool("stats")
		if showStats {
			stats := calculateStats(entries)
			fmt.Println("\nStatistics:")
			fmt.Printf("  Average mood: %.1f\n", stats["average_mood"])
			fmt.Printf("  Average priority: %.1f\n", stats["average_priority"])
			fmt.Printf("  Most common type: %s\n", stats["common_type"])
			fmt.Printf("  Most common tags: %s\n", stats["common_tags"])
		}
	}

	return nil
}

func calculateStats(entries []storage.DailyLogEntry) map[string]interface{} {
	if len(entries) == 0 {
		return map[string]interface{}{}
	}

	// Count by type
	typeCount := make(map[string]int)
	tagCount := make(map[string]int)
	moodSum := 0
	moodCount := 0
	prioritySum := 0
	priorityCount := 0

	for _, entry := range entries {
		typeCount[entry.Type]++

		if entry.Mood > 0 {
			moodSum += entry.Mood
			moodCount++
		}

		if entry.Priority > 0 {
			prioritySum += entry.Priority
			priorityCount++
		}

		for _, tag := range entry.Tags {
			tagCount[tag]++
		}
	}

	// Find most common type
	var commonType string
	maxTypeCount := 0
	for entryType, count := range typeCount {
		if count > maxTypeCount {
			maxTypeCount = count
			commonType = entryType
		}
	}

	// Find most common tags
	var commonTags []string
	maxTagCount := 0
	for tag, count := range tagCount {
		if count > maxTagCount {
			maxTagCount = count
			commonTags = []string{tag}
		} else if count == maxTagCount {
			commonTags = append(commonTags, tag)
		}
	}

	stats := map[string]interface{}{
		"common_type": commonType,
		"common_tags": strings.Join(commonTags, ", "),
	}

	if moodCount > 0 {
		stats["average_mood"] = float64(moodSum) / float64(moodCount)
	} else {
		stats["average_mood"] = 0.0
	}

	if priorityCount > 0 {
		stats["average_priority"] = float64(prioritySum) / float64(priorityCount)
	} else {
		stats["average_priority"] = 0.0
	}

	return stats
}
