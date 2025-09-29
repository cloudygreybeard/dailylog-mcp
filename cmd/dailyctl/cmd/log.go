package cmd

import (
	"fmt"
	"regexp"
	"strconv"
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
	Long: `Create new log entries for daily activities, status updates, notes, or summaries.

Examples:
  dailyctl log activity "Morning meeting with team" --tags work,meeting --status 8
  dailyctl log status "Feeling great today" --status 9 --datetime "yesterday 3pm"
  dailyctl log note "Remember to call mom" --priority 3 --datetime "2 hours ago"
  dailyctl log activity "Completed project" --datetime "2025-09-29 14:30" --status 10
  dailyctl log summary "Productive day overall" --date "2025-09-28"`,
}

var logActivityCmd = &cobra.Command{
	Use:   "activity [title]",
	Short: "Log a daily activity",
	Args:  cobra.ExactArgs(1),
	RunE:  runLogEntry("activity"),
}

var logStatusCmd = &cobra.Command{
	Use:   "status [description]",
	Short: "Log status information",
	Args:  cobra.ExactArgs(1),
	RunE:  runLogEntry("status"),
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
	logCmd.AddCommand(logStatusCmd)
	logCmd.AddCommand(logNoteCmd)
	logCmd.AddCommand(logSummaryCmd)

	// Common flags for all log commands
	addLogFlags := func(cmd *cobra.Command) {
		cmd.Flags().String("date", "", "Date for the entry (YYYY-MM-DD, defaults to today)")
		cmd.Flags().String("datetime", "", "Date and time for the entry (flexible format, e.g. '2025-09-29 14:30', 'yesterday 3pm', '2 hours ago')")
		cmd.Flags().String("description", "", "Detailed description")
		cmd.Flags().StringSlice("tags", []string{}, "Tags for categorization")
		cmd.Flags().Int("status", 0, "Status rating (1-10)")
		cmd.Flags().Int("priority", 0, "Priority level (1-5)")
		cmd.Flags().Int("duration", 0, "Duration in minutes")
		cmd.Flags().String("location", "", "Location")
		
		// Make date and datetime mutually exclusive
		cmd.MarkFlagsMutuallyExclusive("date", "datetime")
	}

	addLogFlags(logActivityCmd)
	addLogFlags(logStatusCmd)
	addLogFlags(logNoteCmd)
	addLogFlags(logSummaryCmd)
}

func runLogEntry(entryType string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		title := args[0]

		// Parse flags
		dateStr, _ := cmd.Flags().GetString("date")
		datetimeStr, _ := cmd.Flags().GetString("datetime")
		description, _ := cmd.Flags().GetString("description")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		status, _ := cmd.Flags().GetInt("status")
		priority, _ := cmd.Flags().GetInt("priority")
		duration, _ := cmd.Flags().GetInt("duration")
		location, _ := cmd.Flags().GetString("location")

		// Parse date/datetime
		var entryDate time.Time
		var err error
		if datetimeStr != "" {
			entryDate, err = parseFlexibleDateTime(datetimeStr)
			if err != nil {
				return fmt.Errorf("invalid datetime format: %s (%v)", datetimeStr, err)
			}
			// Debug output (remove this later)
			if viper.GetBool("verbose") {
				fmt.Printf("DEBUG: Parsed datetime '%s' as %s\n", datetimeStr, entryDate.Format("2006-01-02 15:04:05"))
			}
		} else if dateStr != "" {
			// Parse date and use current time
			dateOnly, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				return fmt.Errorf("invalid date format: %s (use YYYY-MM-DD)", dateStr)
			}
			now := time.Now()
			entryDate = time.Date(dateOnly.Year(), dateOnly.Month(), dateOnly.Day(), 
				now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())
		} else {
			entryDate = time.Now()
		}

		// Validate status range
		if status < 0 || status > 10 {
			return fmt.Errorf("status must be between 1 and 10")
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

		if status > 0 {
			createReq.Status = &status
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
			if entry.Status > 0 {
				fmt.Printf("  Status: %d/10\n", entry.Status)
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

// parseFlexibleDateTime parses various datetime formats, similar to GNU date
func parseFlexibleDateTime(input string) (time.Time, error) {
	input = strings.TrimSpace(input)
	now := time.Now()
	
	// Try common datetime formats first
	formats := []string{
		"2006-01-02 15:04:05",     // YYYY-MM-DD HH:MM:SS
		"2006-01-02 15:04",        // YYYY-MM-DD HH:MM
		"2006-01-02T15:04:05",     // ISO format with T
		"2006-01-02T15:04",        // ISO format with T, no seconds
		"01/02/2006 15:04:05",     // MM/DD/YYYY HH:MM:SS
		"01/02/2006 15:04",        // MM/DD/YYYY HH:MM
		"01/02/2006 3:04 PM",      // MM/DD/YYYY H:MM PM
		"01/02/2006 3:04PM",       // MM/DD/YYYY H:MMPM
		"2006-01-02 3:04 PM",      // YYYY-MM-DD H:MM PM
		"2006-01-02 3:04PM",       // YYYY-MM-DD H:MMPM
		"Jan 2, 2006 15:04",       // Month DD, YYYY HH:MM
		"Jan 2, 2006 3:04 PM",     // Month DD, YYYY H:MM PM
		"2 Jan 2006 15:04",        // DD Month YYYY HH:MM
		"2 Jan 2006 3:04 PM",      // DD Month YYYY H:MM PM
		"15:04",                   // HH:MM (today)
		"3:04 PM",                 // H:MM PM (today)
		"3:04PM",                  // H:MMPM (today)
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, input); err == nil {
			// If no date specified (time only), use today
			if format == "15:04" || format == "3:04 PM" || format == "3:04PM" {
				return time.Date(now.Year(), now.Month(), now.Day(), 
					t.Hour(), t.Minute(), t.Second(), 0, now.Location()), nil
			}
			// For other formats, ensure we use the local timezone
			return time.Date(t.Year(), t.Month(), t.Day(), 
				t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), now.Location()), nil
		}
	}
	
	// Handle relative time expressions
	lower := strings.ToLower(input)
	
	// "now", "today"
	if lower == "now" || lower == "today" {
		return now, nil
	}
	
	// "yesterday", "tomorrow"
	if lower == "yesterday" {
		return now.AddDate(0, 0, -1), nil
	}
	if lower == "tomorrow" {
		return now.AddDate(0, 0, 1), nil
	}
	
	// "X hours ago", "X minutes ago", "X days ago"
	if matched, err := parseRelativeTime(input, now); matched {
		return err, nil
	}
	
	// "yesterday 3pm", "tomorrow 9am"
	if strings.Contains(lower, "yesterday") || strings.Contains(lower, "tomorrow") {
		parts := strings.Fields(lower)
		if len(parts) >= 2 {
			var baseDate time.Time
			if strings.Contains(parts[0], "yesterday") {
				baseDate = now.AddDate(0, 0, -1)
			} else {
				baseDate = now.AddDate(0, 0, 1)
			}
			
			// Try to parse the time part
			timeStr := strings.Join(parts[1:], " ")
			if t, err := parseTimeString(timeStr); err == nil {
				return time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(),
					t.Hour(), t.Minute(), 0, 0, now.Location()), nil
			}
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse datetime: %s", input)
}

// parseRelativeTime handles "X hours ago", "X minutes ago", etc.
func parseRelativeTime(input string, now time.Time) (bool, time.Time) {
	// Regex for "X units ago" or "X units from now"
	re := regexp.MustCompile(`(\d+)\s+(second|minute|hour|day|week|month|year)s?\s+(ago|from\s+now)`)
	matches := re.FindStringSubmatch(strings.ToLower(input))
	
	if len(matches) != 4 {
		return false, time.Time{}
	}
	
	amount, err := strconv.Atoi(matches[1])
	if err != nil {
		return false, time.Time{}
	}
	
	unit := matches[2]
	direction := matches[3]
	
	if direction == "ago" {
		amount = -amount
	}
	
	switch unit {
	case "second":
		return true, now.Add(time.Duration(amount) * time.Second)
	case "minute":
		return true, now.Add(time.Duration(amount) * time.Minute)
	case "hour":
		return true, now.Add(time.Duration(amount) * time.Hour)
	case "day":
		return true, now.AddDate(0, 0, amount)
	case "week":
		return true, now.AddDate(0, 0, amount*7)
	case "month":
		return true, now.AddDate(0, amount, 0)
	case "year":
		return true, now.AddDate(amount, 0, 0)
	}
	
	return false, time.Time{}
}

// parseTimeString parses time strings like "3pm", "14:30", "9:15am"
func parseTimeString(input string) (time.Time, error) {
	timeFormats := []string{
		"15:04",     // 24-hour
		"3:04 PM",   // 12-hour with space
		"3:04PM",    // 12-hour without space
		"3PM",       // hour only with PM
		"15",        // hour only 24-hour
	}
	
	for _, format := range timeFormats {
		if t, err := time.Parse(format, input); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse time: %s", input)
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
