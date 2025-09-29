package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"dailylog/internal/storage"
)

// summarizeCmd represents the summarize command
var summarizeCmd = &cobra.Command{
	Use:   "summarize",
	Short: "Generate summaries of log entries",
	Long: `Generate summaries for daily, weekly, monthly, or custom periods.

Examples:
  dailyctl summarize day
  dailyctl summarize week
  dailyctl summarize month
  dailyctl summarize day --date 2025-09-29
  dailyctl summarize custom --date-start 2025-09-01 --date-end 2025-09-30`,
}

var summarizeDayCmd = &cobra.Command{
	Use:   "day",
	Short: "Summarize a day",
	RunE:  runSummarize("day"),
}

var summarizeWeekCmd = &cobra.Command{
	Use:   "week",
	Short: "Summarize a week",
	RunE:  runSummarize("week"),
}

var summarizeMonthCmd = &cobra.Command{
	Use:   "month",
	Short: "Summarize a month",
	RunE:  runSummarize("month"),
}

var summarizeCustomCmd = &cobra.Command{
	Use:   "custom",
	Short: "Summarize a custom date range",
	RunE:  runSummarize("custom"),
}

func init() {
	rootCmd.AddCommand(summarizeCmd)

	// Add subcommands
	summarizeCmd.AddCommand(summarizeDayCmd)
	summarizeCmd.AddCommand(summarizeWeekCmd)
	summarizeCmd.AddCommand(summarizeMonthCmd)
	summarizeCmd.AddCommand(summarizeCustomCmd)

	// Add flags
	addSummaryFlags := func(cmd *cobra.Command) {
		cmd.Flags().String("date", "", "Date for summary (YYYY-MM-DD, defaults to today)")
		cmd.Flags().String("date-start", "", "Start date for custom range (YYYY-MM-DD)")
		cmd.Flags().String("date-end", "", "End date for custom range (YYYY-MM-DD)")
		cmd.Flags().Bool("ai", false, "Use AI for enhanced summary generation")
		cmd.Flags().String("prompt", "", "Custom prompt for AI summary")
		cmd.Flags().Bool("save", false, "Save summary to the log data")
	}

	addSummaryFlags(summarizeDayCmd)
	addSummaryFlags(summarizeWeekCmd)
	addSummaryFlags(summarizeMonthCmd)
	addSummaryFlags(summarizeCustomCmd)
}

func runSummarize(summaryType string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Parse flags
		dateStr, _ := cmd.Flags().GetString("date")
		dateStartStr, _ := cmd.Flags().GetString("date-start")
		dateEndStr, _ := cmd.Flags().GetString("date-end")
		useAI, _ := cmd.Flags().GetBool("ai")
		prompt, _ := cmd.Flags().GetString("prompt")
		save, _ := cmd.Flags().GetBool("save")

		// Parse target date
		var targetDate time.Time
		var err error
		if dateStr != "" {
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

		// Build summary request
		summaryReq := storage.SummaryRequest{
			Type:   summaryType,
			Date:   targetDate,
			UseAI:  useAI,
			Prompt: prompt,
		}

		// Handle custom date range
		if summaryType == "custom" {
			if dateStartStr == "" || dateEndStr == "" {
				return fmt.Errorf("custom summary requires both --date-start and --date-end")
			}

			startDate, err1 := time.Parse("2006-01-02", dateStartStr)
			endDate, err2 := time.Parse("2006-01-02", dateEndStr)
			if err1 != nil || err2 != nil {
				return fmt.Errorf("invalid date format in range")
			}

			if startDate.After(endDate) {
				return fmt.Errorf("start date cannot be after end date")
			}

			summaryReq.StartDate = &startDate
			summaryReq.EndDate = &endDate
		}

		// Generate summary
		summaryResult, err := storageProvider.GenerateSummary(summaryReq)
		if err != nil {
			return fmt.Errorf("failed to generate summary: %v", err)
		}

		// Save summary if requested
		if save {
			err = storageProvider.SaveSummary(summaryResult, summaryType, targetDate)
			if err != nil {
				fmt.Printf("Warning: Failed to save summary: %v\n", err)
			} else {
				fmt.Println("✓ Summary saved to log data")
			}
		}

		// Output summary
		return outputSummary(summaryResult)
	}
}

func outputSummary(summary *storage.SummaryResponse) error {
	fmt.Printf("📊 %s Summary - %s\n",
		strings.Title(summary.Type), summary.Period)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	// Main summary content
	fmt.Println(summary.Summary)
	fmt.Println()

	// Statistics
	if len(summary.Stats) > 0 {
		fmt.Println("📈 Statistics:")
		fmt.Println(strings.Repeat("-", 20))

		if totalEntries, ok := summary.Stats["total_entries"].(int); ok {
			fmt.Printf("  Total entries: %d\n", totalEntries)
		}
		if totalDays, ok := summary.Stats["total_days"].(int); ok {
			fmt.Printf("  Total days: %d\n", totalDays)
		}
		if avgStatus, ok := summary.Stats["average_status"].(float64); ok && avgStatus > 0 {
			fmt.Printf("  Average status: %.1f/10\n", avgStatus)
		}
		if entriesPerDay, ok := summary.Stats["entries_per_day"].(float64); ok {
			fmt.Printf("  Entries per day: %.1f\n", entriesPerDay)
		}

		fmt.Println()
	}

	fmt.Printf("Generated at: %s\n", summary.CreatedAt.Format("2006-01-02 15:04:05"))

	return nil
}
