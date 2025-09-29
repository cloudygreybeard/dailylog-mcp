package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"dailylog/internal/storage"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search through log entries",
	Long: `Search through log entries by text, tags, mood, or other criteria.

Examples:
  dailyctl search --query "exercise"
  dailyctl search --tags work,meeting
  dailyctl search --mood-min 8 --mood-max 10
  dailyctl search --query "project" --type activity --date-start 2025-09-01`,
	RunE: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Add search flags
	searchCmd.Flags().String("query", "", "Search text in titles and descriptions")
	searchCmd.Flags().String("date-start", "", "Start date for search range (YYYY-MM-DD)")
	searchCmd.Flags().String("date-end", "", "End date for search range (YYYY-MM-DD)")
	searchCmd.Flags().String("type", "", "Filter by entry type")
	searchCmd.Flags().StringSlice("tags", []string{}, "Filter by tags")
	searchCmd.Flags().Int("mood-min", 0, "Minimum mood rating")
	searchCmd.Flags().Int("mood-max", 0, "Maximum mood rating")
	searchCmd.Flags().Int("limit", 50, "Maximum number of results")
}

func runSearch(cmd *cobra.Command, args []string) error {
	// Get search parameters
	query, _ := cmd.Flags().GetString("query")
	dateStartStr, _ := cmd.Flags().GetString("date-start")
	dateEndStr, _ := cmd.Flags().GetString("date-end")
	entryType, _ := cmd.Flags().GetString("type")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	moodMin, _ := cmd.Flags().GetInt("mood-min")
	moodMax, _ := cmd.Flags().GetInt("mood-max")
	limit, _ := cmd.Flags().GetInt("limit")

	// Validate that at least one search criterion is provided
	if query == "" && entryType == "" && len(tags) == 0 && moodMin == 0 && moodMax == 0 {
		return fmt.Errorf("at least one search criterion must be provided")
	}

	// Parse dates
	var dateStart, dateEnd *time.Time
	if dateStartStr != "" {
		start, err := time.Parse("2006-01-02", dateStartStr)
		if err != nil {
			return fmt.Errorf("invalid start date format: %s (use YYYY-MM-DD)", dateStartStr)
		}
		dateStart = &start
	}
	if dateEndStr != "" {
		end, err := time.Parse("2006-01-02", dateEndStr)
		if err != nil {
			return fmt.Errorf("invalid end date format: %s (use YYYY-MM-DD)", dateEndStr)
		}
		dateEnd = &end
	}

	// Validate mood range
	if moodMin < 0 || moodMin > 10 {
		return fmt.Errorf("mood-min must be between 1 and 10")
	}
	if moodMax < 0 || moodMax > 10 {
		return fmt.Errorf("mood-max must be between 1 and 10")
	}
	if moodMin > 0 && moodMax > 0 && moodMin > moodMax {
		return fmt.Errorf("mood-min cannot be greater than mood-max")
	}

	// Create storage provider
	storageProvider, err := createStorageProvider()
	if err != nil {
		return fmt.Errorf("failed to create storage provider: %v", err)
	}

	// Build search request
	searchReq := storage.LogSearchRequest{
		SearchText: query,
		DateStart:  dateStart,
		DateEnd:    dateEnd,
		Type:       entryType,
		Tags:       tags,
		Limit:      limit,
	}

	if moodMin > 0 {
		searchReq.MoodMin = &moodMin
	}
	if moodMax > 0 {
		searchReq.MoodMax = &moodMax
	}

	// Perform search
	searchResult, err := storageProvider.SearchLogs(searchReq)
	if err != nil {
		return fmt.Errorf("search failed: %v", err)
	}

	// Output results
	return outputSearchResults(searchResult, query)
}

func outputSearchResults(result *storage.LogSearchResponse, query string) error {
	fmt.Printf("Search Results")
	if query != "" {
		fmt.Printf(" for '%s'", query)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	if len(result.Entries) == 0 {
		fmt.Println("No entries found matching the search criteria.")
		return nil
	}

	// Group entries by date for better readability
	entriesByDate := make(map[string][]storage.DailyLogEntry)
	for _, entry := range result.Entries {
		dateKey := entry.Timestamp.Format("2006-01-02")
		entriesByDate[dateKey] = append(entriesByDate[dateKey], entry)
	}

	// Sort dates and display
	var dates []string
	for date := range entriesByDate {
		dates = append(dates, date)
	}

	// Simple sort (for more robust sorting, would use sort package)
	for i := 0; i < len(dates); i++ {
		for j := i + 1; j < len(dates); j++ {
			if dates[i] > dates[j] {
				dates[i], dates[j] = dates[j], dates[i]
			}
		}
	}

	for _, date := range dates {
		entries := entriesByDate[date]
		fmt.Printf("📅 %s (%d entries)\n", date, len(entries))
		fmt.Println(strings.Repeat("-", 30))

		for _, entry := range entries {
			fmt.Printf("  🕐 %s - %s [%s]\n",
				entry.Timestamp.Format("15:04"), entry.Title, entry.Type)

			if entry.Description != "" {
				fmt.Printf("     %s\n", entry.Description)
			}

			// Show metadata
			var metadata []string
			if len(entry.Tags) > 0 {
				metadata = append(metadata, fmt.Sprintf("Tags: %s", strings.Join(entry.Tags, ", ")))
			}
			if entry.Mood > 0 {
				metadata = append(metadata, fmt.Sprintf("Mood: %d/10", entry.Mood))
			}
			if entry.Priority > 0 {
				metadata = append(metadata, fmt.Sprintf("Priority: %d/5", entry.Priority))
			}
			if entry.Duration != nil && *entry.Duration > 0 {
				metadata = append(metadata, fmt.Sprintf("Duration: %dm", *entry.Duration))
			}
			if entry.Location != "" {
				metadata = append(metadata, fmt.Sprintf("Location: %s", entry.Location))
			}

			if len(metadata) > 0 {
				fmt.Printf("     %s\n", strings.Join(metadata, " | "))
			}

			fmt.Println()
		}
	}

	fmt.Printf("Found %d entries total\n", result.TotalCount)
	return nil
}
