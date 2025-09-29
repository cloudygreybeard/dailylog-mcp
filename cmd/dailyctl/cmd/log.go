package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"dailylog/internal/providers"
	"dailylog/internal/storage"
)

// logCmd represents the log command
var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Create log entries",
	Long: `Create new log entries for daily activities, moods, notes, or summaries.

Examples:
  dailyctl log activity "Morning meeting with team" --tags work,meeting --mood 8
  dailyctl log mood "Feeling great today" --mood 9
  dailyctl log note "Remember to call mom" --priority 3
  dailyctl log summary "Productive day overall"`,
}

var logActivityCmd = &cobra.Command{
	Use:   "activity [title]",
	Short: "Log a daily activity",
	Args:  cobra.ExactArgs(1),
	RunE:  runLogEntry("activity"),
}

var logMoodCmd = &cobra.Command{
	Use:   "mood [description]",
	Short: "Log mood information",
	Args:  cobra.ExactArgs(1),
	RunE:  runLogEntry("mood"),
}

var logNoteCmd = &cobra.Command{
	Use:   "note [content]",
	Short: "Create a note entry",
	Args:  cobra.ExactArgs(1),
	RunE:  runLogEntry("note"),
}

var logSummaryCmd = &cobra.Command{
	Use:   "summary [content]",
	Short: "Create a summary entry",
	Args:  cobra.ExactArgs(1),
	RunE:  runLogEntry("summary"),
}

func init() {
	rootCmd.AddCommand(logCmd)

	// Add subcommands
	logCmd.AddCommand(logActivityCmd)
	logCmd.AddCommand(logMoodCmd)
	logCmd.AddCommand(logNoteCmd)
	logCmd.AddCommand(logSummaryCmd)

	// Common flags for all log commands
	addLogFlags := func(cmd *cobra.Command) {
		cmd.Flags().String("date", "", "Date for the entry (YYYY-MM-DD, defaults to today)")
		cmd.Flags().String("description", "", "Detailed description")
		cmd.Flags().StringSlice("tags", []string{}, "Tags for categorization")
		cmd.Flags().Int("mood", 0, "Mood rating (1-10)")
		cmd.Flags().Int("priority", 0, "Priority level (1-5)")
		cmd.Flags().Int("duration", 0, "Duration in minutes")
		cmd.Flags().String("location", "", "Location")
	}

	addLogFlags(logActivityCmd)
	addLogFlags(logMoodCmd)
	addLogFlags(logNoteCmd)
	addLogFlags(logSummaryCmd)
}

func runLogEntry(entryType string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		title := args[0]

		// Parse flags
		dateStr, _ := cmd.Flags().GetString("date")
		description, _ := cmd.Flags().GetString("description")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		mood, _ := cmd.Flags().GetInt("mood")
		priority, _ := cmd.Flags().GetInt("priority")
		duration, _ := cmd.Flags().GetInt("duration")
		location, _ := cmd.Flags().GetString("location")

		// Parse date
		var entryDate time.Time
		var err error
		if dateStr != "" {
			entryDate, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				return fmt.Errorf("invalid date format: %s (use YYYY-MM-DD)", dateStr)
			}
		} else {
			entryDate = time.Now()
		}

		// Validate mood range
		if mood < 0 || mood > 10 {
			return fmt.Errorf("mood must be between 1 and 10")
		}

		// Validate priority range
		if priority < 0 || priority > 5 {
			return fmt.Errorf("priority must be between 1 and 5")
		}

		// Create storage provider
		storageProvider, err := createStorageProvider()
		if err != nil {
			return fmt.Errorf("failed to create storage provider: %v", err)
		}

		// Create the log entry
		createReq := storage.CreateLogEntryRequest{
			Date:        entryDate,
			Type:        entryType,
			Title:       title,
			Description: description,
			Tags:        tags,
			Location:    location,
		}

		if mood > 0 {
			createReq.Mood = &mood
		}
		if priority > 0 {
			createReq.Priority = &priority
		}
		if duration > 0 {
			createReq.Duration = &duration
		}

		entry, err := storageProvider.CreateEntry(createReq)
		if err != nil {
			return fmt.Errorf("failed to create entry: %v", err)
		}

		// Output result
		outputFormat := viper.GetString("output.format")
		switch outputFormat {
		case "json":
			return outputJSON(entry)
		case "yaml":
			return outputYAML(entry)
		default:
			fmt.Printf("✓ Created %s entry: %s\n", entryType, entry.Title)
			fmt.Printf("  ID: %s\n", entry.ID)
			fmt.Printf("  Date: %s\n", entryDate.Format("2006-01-02"))
			fmt.Printf("  Time: %s\n", entry.Timestamp.Format("15:04:05"))
			if len(entry.Tags) > 0 {
				fmt.Printf("  Tags: %s\n", strings.Join(entry.Tags, ", "))
			}
			if entry.Mood > 0 {
				fmt.Printf("  Mood: %d/10\n", entry.Mood)
			}
			if entry.Priority > 0 {
				fmt.Printf("  Priority: %d/5\n", entry.Priority)
			}
			if entry.Duration != nil && *entry.Duration > 0 {
				fmt.Printf("  Duration: %d minutes\n", *entry.Duration)
			}
			if entry.Location != "" {
				fmt.Printf("  Location: %s\n", entry.Location)
			}
		}

		return nil
	}
}

func createStorageProvider() (storage.DailyLogStorage, error) {
	config := storage.Config{
		StorageType: "github",
		GitHubRepo:  viper.GetString("github.repo"),
		GitHubToken: viper.GetString("github.token"),
		GitHubPath:  viper.GetString("github.path"),
	}

	if config.GitHubRepo == "" {
		return nil, fmt.Errorf("GitHub repository not configured (use --github-repo or set DAILYLOG_GITHUB_REPO)")
	}
	if config.GitHubToken == "" {
		return nil, fmt.Errorf("GitHub token not configured (use --github-token or set DAILYLOG_GITHUB_TOKEN)")
	}

	return providers.NewGitHubStorageProvider(config)
}
