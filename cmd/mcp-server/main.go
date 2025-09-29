package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"dailylog/internal/providers"
	"dailylog/internal/storage"
)

// Version information (set by build)
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// Server holds our daily log implementation
type Server struct {
	storage storage.DailyLogStorage
}

// === MCP INPUT/OUTPUT TYPES ===

// LogEntryInput defines parameters for creating a log entry
type LogEntryInput struct {
	Date        string            `json:"date,omitempty" jsonschema:"Date in YYYY-MM-DD format (defaults to today)"`
	Type        string            `json:"type" jsonschema:"Entry type: activity, mood, note, summary"`
	Title       string            `json:"title" jsonschema:"Entry title"`
	Description string            `json:"description" jsonschema:"Entry description"`
	Tags        []string          `json:"tags,omitempty" jsonschema:"Tags for categorization"`
	Mood        *int              `json:"mood,omitempty" jsonschema:"Mood rating 1-10"`
	Priority    *int              `json:"priority,omitempty" jsonschema:"Priority 1-5"`
	Duration    *int              `json:"duration,omitempty" jsonschema:"Duration in minutes"`
	Location    string            `json:"location,omitempty" jsonschema:"Location"`
	Metadata    map[string]string `json:"metadata,omitempty" jsonschema:"Additional metadata"`
}

// LogEntryOutput defines the response for log entry operations
type LogEntryOutput struct {
	ID          string            `json:"id" jsonschema:"Entry ID"`
	Date        string            `json:"date" jsonschema:"Entry date"`
	Timestamp   string            `json:"timestamp" jsonschema:"Entry timestamp"`
	Type        string            `json:"type" jsonschema:"Entry type"`
	Title       string            `json:"title" jsonschema:"Entry title"`
	Description string            `json:"description" jsonschema:"Entry description"`
	Tags        []string          `json:"tags,omitempty" jsonschema:"Entry tags"`
	Mood        int               `json:"mood,omitempty" jsonschema:"Mood rating"`
	Priority    int               `json:"priority,omitempty" jsonschema:"Priority"`
	Duration    *int              `json:"duration,omitempty" jsonschema:"Duration in minutes"`
	Location    string            `json:"location,omitempty" jsonschema:"Location"`
	Metadata    map[string]string `json:"metadata,omitempty" jsonschema:"Metadata"`
	Success     bool              `json:"success" jsonschema:"Whether operation was successful"`
	Message     string            `json:"message,omitempty" jsonschema:"Success or error message"`
}

// GetEntriesInput defines parameters for retrieving log entries
type GetEntriesInput struct {
	Date         string   `json:"date,omitempty" jsonschema:"Specific date in YYYY-MM-DD format"`
	DateStart    string   `json:"date_start,omitempty" jsonschema:"Start date for range query"`
	DateEnd      string   `json:"date_end,omitempty" jsonschema:"End date for range query"`
	Type         string   `json:"type,omitempty" jsonschema:"Filter by entry type"`
	Tags         []string `json:"tags,omitempty" jsonschema:"Filter by tags"`
	Limit        int      `json:"limit,omitempty" jsonschema:"Maximum number of entries to return"`
	IncludeStats bool     `json:"include_stats,omitempty" jsonschema:"Include summary statistics"`
}

// GetEntriesOutput defines the response for getting entries
type GetEntriesOutput struct {
	Entries    []LogEntryOutput `json:"entries" jsonschema:"Log entries"`
	TotalCount int              `json:"total_count" jsonschema:"Total number of entries found"`
	Stats      map[string]any   `json:"stats,omitempty" jsonschema:"Summary statistics"`
	Period     string           `json:"period,omitempty" jsonschema:"Time period covered"`
	Success    bool             `json:"success" jsonschema:"Whether operation was successful"`
	Message    string           `json:"message,omitempty" jsonschema:"Success or error message"`
}

// SearchLogsInput defines parameters for searching logs
type SearchLogsInput struct {
	Query     string   `json:"query,omitempty" jsonschema:"Search text in titles and descriptions"`
	DateStart string   `json:"date_start,omitempty" jsonschema:"Start date for search range"`
	DateEnd   string   `json:"date_end,omitempty" jsonschema:"End date for search range"`
	Type      string   `json:"type,omitempty" jsonschema:"Filter by entry type"`
	Tags      []string `json:"tags,omitempty" jsonschema:"Filter by tags"`
	MoodMin   *int     `json:"mood_min,omitempty" jsonschema:"Minimum mood rating"`
	MoodMax   *int     `json:"mood_max,omitempty" jsonschema:"Maximum mood rating"`
	Limit     int      `json:"limit,omitempty" jsonschema:"Maximum number of results"`
}

// SearchLogsOutput defines the response for searching logs
type SearchLogsOutput struct {
	Entries     []LogEntryOutput `json:"entries" jsonschema:"Matching log entries"`
	TotalCount  int              `json:"total_count" jsonschema:"Total number of matches"`
	SearchQuery string           `json:"search_query,omitempty" jsonschema:"The search query used"`
	Success     bool             `json:"success" jsonschema:"Whether operation was successful"`
	Message     string           `json:"message,omitempty" jsonschema:"Success or error message"`
}

// SummarizePeriodInput defines parameters for generating summaries
type SummarizePeriodInput struct {
	Type      string `json:"type" jsonschema:"Summary type: day, week, month"`
	Date      string `json:"date,omitempty" jsonschema:"Date for summary (defaults to today)"`
	DateStart string `json:"date_start,omitempty" jsonschema:"Start date for custom range"`
	DateEnd   string `json:"date_end,omitempty" jsonschema:"End date for custom range"`
	UseAI     bool   `json:"use_ai,omitempty" jsonschema:"Use AI for enhanced summary generation"`
	Prompt    string `json:"prompt,omitempty" jsonschema:"Custom prompt for AI summary"`
}

// SummarizePeriodOutput defines the response for summary generation
type SummarizePeriodOutput struct {
	Summary   string         `json:"summary" jsonschema:"Generated summary"`
	Type      string         `json:"type" jsonschema:"Summary type"`
	Period    string         `json:"period" jsonschema:"Time period summarized"`
	Stats     map[string]any `json:"stats" jsonschema:"Statistical information"`
	Timestamp string         `json:"timestamp" jsonschema:"When summary was generated"`
	Success   bool           `json:"success" jsonschema:"Whether operation was successful"`
	Message   string         `json:"message,omitempty" jsonschema:"Success or error message"`
}

// AIAssistInput defines parameters for AI assistance features
type AIAssistInput struct {
	Action string `json:"action" jsonschema:"AI action: improve_wording, suggest_tags, analyze_mood, generate_insights"`
	Text   string `json:"text,omitempty" jsonschema:"Text to improve or analyze"`
	Date   string `json:"date,omitempty" jsonschema:"Date for context (for analysis actions)"`
}

// AIAssistOutput defines the response for AI assistance
type AIAssistOutput struct {
	Result      string   `json:"result" jsonschema:"AI-generated result"`
	Action      string   `json:"action" jsonschema:"Action performed"`
	Suggestions []string `json:"suggestions,omitempty" jsonschema:"Additional suggestions"`
	Success     bool     `json:"success" jsonschema:"Whether operation was successful"`
	Message     string   `json:"message,omitempty" jsonschema:"Success or error message"`
}

// === TOOL IMPLEMENTATIONS ===

// LogEntry implements the dailylog_entry tool
func (s *Server) LogEntry(ctx context.Context, req *mcp.CallToolRequest, input LogEntryInput) (
	*mcp.CallToolResult,
	LogEntryOutput,
	error,
) {
	log.Printf("LogEntry called with input: %+v", input)

	// Parse date
	var entryDate time.Time
	var err error
	if input.Date != "" {
		entryDate, err = time.Parse("2006-01-02", input.Date)
		if err != nil {
			return nil, LogEntryOutput{
				Success: false,
				Message: fmt.Sprintf("Invalid date format: %s", input.Date),
			}, nil
		}
	} else {
		entryDate = time.Now()
	}

	// Validate required fields
	if input.Type == "" {
		return nil, LogEntryOutput{
			Success: false,
			Message: "Entry type is required",
		}, nil
	}
	if input.Title == "" {
		return nil, LogEntryOutput{
			Success: false,
			Message: "Entry title is required",
		}, nil
	}

	// Create the log entry
	createReq := storage.CreateLogEntryRequest{
		Date:        entryDate,
		Type:        input.Type,
		Title:       input.Title,
		Description: input.Description,
		Tags:        input.Tags,
		Mood:        input.Mood,
		Priority:    input.Priority,
		Duration:    input.Duration,
		Location:    input.Location,
		Metadata:    input.Metadata,
	}

	entry, err := s.storage.CreateEntry(createReq)
	if err != nil {
		return nil, LogEntryOutput{
			Success: false,
			Message: fmt.Sprintf("Failed to create entry: %v", err),
		}, nil
	}

	result := LogEntryOutput{
		ID:          entry.ID,
		Date:        entryDate.Format("2006-01-02"),
		Timestamp:   entry.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		Type:        entry.Type,
		Title:       entry.Title,
		Description: entry.Description,
		Tags:        entry.Tags,
		Mood:        entry.Mood,
		Priority:    entry.Priority,
		Duration:    entry.Duration,
		Location:    entry.Location,
		Metadata:    entry.Metadata,
		Success:     true,
		Message:     fmt.Sprintf("Entry '%s' created successfully", entry.Title),
	}

	return nil, result, nil
}

// GetEntries implements the dailylog_get_entries tool
func (s *Server) GetEntries(ctx context.Context, req *mcp.CallToolRequest, input GetEntriesInput) (
	*mcp.CallToolResult,
	GetEntriesOutput,
	error,
) {
	log.Printf("GetEntries called with input: %+v", input)

	var entries []storage.DailyLogEntry
	var err error
	var period string

	if input.Date != "" {
		// Get entries for a specific date
		date, parseErr := time.Parse("2006-01-02", input.Date)
		if parseErr != nil {
			return nil, GetEntriesOutput{
				Success: false,
				Message: fmt.Sprintf("Invalid date format: %s", input.Date),
			}, nil
		}

		dayLog, err := s.storage.GetDay(date)
		if err != nil {
			return nil, GetEntriesOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to get day: %v", err),
			}, nil
		}

		entries = dayLog.Entries
		period = input.Date

	} else if input.DateStart != "" && input.DateEnd != "" {
		// Get entries for a date range
		startDate, err1 := time.Parse("2006-01-02", input.DateStart)
		endDate, err2 := time.Parse("2006-01-02", input.DateEnd)
		if err1 != nil || err2 != nil {
			return nil, GetEntriesOutput{
				Success: false,
				Message: "Invalid date format in range",
			}, nil
		}

		searchReq := storage.LogSearchRequest{
			DateStart: &startDate,
			DateEnd:   &endDate,
			Type:      input.Type,
			Tags:      input.Tags,
			Limit:     input.Limit,
		}

		searchResult, err := s.storage.SearchLogs(searchReq)
		if err != nil {
			return nil, GetEntriesOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to search logs: %v", err),
			}, nil
		}

		entries = searchResult.Entries
		period = fmt.Sprintf("%s to %s", input.DateStart, input.DateEnd)

	} else {
		// Get today's entries by default
		today := time.Now()
		dayLog, err := s.storage.GetDay(today)
		if err != nil {
			return nil, GetEntriesOutput{
				Success: false,
				Message: fmt.Sprintf("Failed to get today's entries: %v", err),
			}, nil
		}

		entries = dayLog.Entries
		period = today.Format("2006-01-02")
	}

	// Convert to output format
	outputEntries := make([]LogEntryOutput, 0, len(entries))
	for _, entry := range entries {
		outputEntry := LogEntryOutput{
			ID:          entry.ID,
			Date:        entry.Timestamp.Format("2006-01-02"),
			Timestamp:   entry.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			Type:        entry.Type,
			Title:       entry.Title,
			Description: entry.Description,
			Tags:        entry.Tags,
			Mood:        entry.Mood,
			Priority:    entry.Priority,
			Duration:    entry.Duration,
			Location:    entry.Location,
			Metadata:    entry.Metadata,
			Success:     true,
		}
		outputEntries = append(outputEntries, outputEntry)
	}

	result := GetEntriesOutput{
		Entries:    outputEntries,
		TotalCount: len(outputEntries),
		Period:     period,
		Success:    true,
		Message:    fmt.Sprintf("Found %d entries", len(outputEntries)),
	}

	// Add stats if requested
	if input.IncludeStats {
		result.Stats = s.calculateEntryStats(entries)
	}

	return nil, result, nil
}

// SearchLogs implements the dailylog_search tool
func (s *Server) SearchLogs(ctx context.Context, req *mcp.CallToolRequest, input SearchLogsInput) (
	*mcp.CallToolResult,
	SearchLogsOutput,
	error,
) {
	log.Printf("SearchLogs called with input: %+v", input)

	// Build search request
	searchReq := storage.LogSearchRequest{
		SearchText: input.Query,
		Type:       input.Type,
		Tags:       input.Tags,
		MoodMin:    input.MoodMin,
		MoodMax:    input.MoodMax,
		Limit:      input.Limit,
	}

	// Parse date range if provided
	if input.DateStart != "" {
		startDate, err := time.Parse("2006-01-02", input.DateStart)
		if err != nil {
			return nil, SearchLogsOutput{
				Success: false,
				Message: fmt.Sprintf("Invalid start date format: %s", input.DateStart),
			}, nil
		}
		searchReq.DateStart = &startDate
	}

	if input.DateEnd != "" {
		endDate, err := time.Parse("2006-01-02", input.DateEnd)
		if err != nil {
			return nil, SearchLogsOutput{
				Success: false,
				Message: fmt.Sprintf("Invalid end date format: %s", input.DateEnd),
			}, nil
		}
		searchReq.DateEnd = &endDate
	}

	// Perform search
	searchResult, err := s.storage.SearchLogs(searchReq)
	if err != nil {
		return nil, SearchLogsOutput{
			Success: false,
			Message: fmt.Sprintf("Search failed: %v", err),
		}, nil
	}

	// Convert to output format
	outputEntries := make([]LogEntryOutput, 0, len(searchResult.Entries))
	for _, entry := range searchResult.Entries {
		outputEntry := LogEntryOutput{
			ID:          entry.ID,
			Date:        entry.Timestamp.Format("2006-01-02"),
			Timestamp:   entry.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			Type:        entry.Type,
			Title:       entry.Title,
			Description: entry.Description,
			Tags:        entry.Tags,
			Mood:        entry.Mood,
			Priority:    entry.Priority,
			Duration:    entry.Duration,
			Location:    entry.Location,
			Metadata:    entry.Metadata,
			Success:     true,
		}
		outputEntries = append(outputEntries, outputEntry)
	}

	result := SearchLogsOutput{
		Entries:     outputEntries,
		TotalCount:  searchResult.TotalCount,
		SearchQuery: input.Query,
		Success:     true,
		Message:     fmt.Sprintf("Found %d matching entries", len(outputEntries)),
	}

	return nil, result, nil
}

// SummarizePeriod implements the dailylog_summarize tool
func (s *Server) SummarizePeriod(ctx context.Context, req *mcp.CallToolRequest, input SummarizePeriodInput) (
	*mcp.CallToolResult,
	SummarizePeriodOutput,
	error,
) {
	log.Printf("SummarizePeriod called with input: %+v", input)

	// Parse date
	var targetDate time.Time
	var err error
	if input.Date != "" {
		targetDate, err = time.Parse("2006-01-02", input.Date)
		if err != nil {
			return nil, SummarizePeriodOutput{
				Success: false,
				Message: fmt.Sprintf("Invalid date format: %s", input.Date),
			}, nil
		}
	} else {
		targetDate = time.Now()
	}

	// Create summary request
	summaryReq := storage.SummaryRequest{
		Type:   input.Type,
		Date:   targetDate,
		UseAI:  input.UseAI,
		Prompt: input.Prompt,
	}

	// Handle custom date range
	if input.DateStart != "" && input.DateEnd != "" {
		startDate, err1 := time.Parse("2006-01-02", input.DateStart)
		endDate, err2 := time.Parse("2006-01-02", input.DateEnd)
		if err1 != nil || err2 != nil {
			return nil, SummarizePeriodOutput{
				Success: false,
				Message: "Invalid date format in range",
			}, nil
		}
		summaryReq.StartDate = &startDate
		summaryReq.EndDate = &endDate
		summaryReq.Type = "custom"
	}

	// Generate summary
	summaryResult, err := s.storage.GenerateSummary(summaryReq)
	if err != nil {
		return nil, SummarizePeriodOutput{
			Success: false,
			Message: fmt.Sprintf("Failed to generate summary: %v", err),
		}, nil
	}

	result := SummarizePeriodOutput{
		Summary:   summaryResult.Summary,
		Type:      summaryResult.Type,
		Period:    summaryResult.Period,
		Stats:     summaryResult.Stats,
		Timestamp: summaryResult.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Success:   true,
		Message:   fmt.Sprintf("Summary generated for %s", summaryResult.Period),
	}

	return nil, result, nil
}

// AIAssist implements the dailylog_ai_assist tool
func (s *Server) AIAssist(ctx context.Context, req *mcp.CallToolRequest, input AIAssistInput) (
	*mcp.CallToolResult,
	AIAssistOutput,
	error,
) {
	log.Printf("AIAssist called with input: %+v", input)

	// Basic implementation - would integrate with actual AI services
	var result string
	var suggestions []string

	switch input.Action {
	case "improve_wording":
		if input.Text == "" {
			return nil, AIAssistOutput{
				Success: false,
				Message: "Text is required for improve_wording action",
			}, nil
		}
		result = s.improveWording(input.Text)

	case "suggest_tags":
		if input.Text == "" {
			return nil, AIAssistOutput{
				Success: false,
				Message: "Text is required for suggest_tags action",
			}, nil
		}
		suggestions = s.suggestTags(input.Text)
		result = fmt.Sprintf("Suggested tags: %s", strings.Join(suggestions, ", "))

	case "analyze_mood":
		if input.Date == "" {
			input.Date = time.Now().Format("2006-01-02")
		}
		result = s.analyzeMood(input.Date)

	case "generate_insights":
		if input.Date == "" {
			input.Date = time.Now().Format("2006-01-02")
		}
		result = s.generateInsights(input.Date)

	default:
		return nil, AIAssistOutput{
			Success: false,
			Message: fmt.Sprintf("Unknown AI action: %s", input.Action),
		}, nil
	}

	output := AIAssistOutput{
		Result:      result,
		Action:      input.Action,
		Suggestions: suggestions,
		Success:     true,
		Message:     fmt.Sprintf("AI %s completed successfully", input.Action),
	}

	return nil, output, nil
}

// === HELPER METHODS ===

func (s *Server) calculateEntryStats(entries []storage.DailyLogEntry) map[string]any {
	stats := map[string]any{
		"total_entries": len(entries),
	}

	if len(entries) == 0 {
		return stats
	}

	// Count by type
	typeCount := make(map[string]int)
	moodSum := 0
	moodCount := 0
	prioritySum := 0
	priorityCount := 0
	tagCount := make(map[string]int)

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

	stats["by_type"] = typeCount
	stats["top_tags"] = tagCount

	if moodCount > 0 {
		stats["average_mood"] = float64(moodSum) / float64(moodCount)
	}

	if priorityCount > 0 {
		stats["average_priority"] = float64(prioritySum) / float64(priorityCount)
	}

	return stats
}

// Basic AI simulation methods (would be replaced with actual AI integration)
func (s *Server) improveWording(text string) string {
	// Placeholder implementation
	return fmt.Sprintf("Enhanced: %s", text)
}

func (s *Server) suggestTags(text string) []string {
	// Placeholder implementation - basic keyword extraction
	words := strings.Fields(strings.ToLower(text))
	tags := []string{}

	commonTags := map[string]bool{
		"work": true, "meeting": true, "exercise": true, "meal": true,
		"family": true, "friends": true, "health": true, "learning": true,
	}

	for _, word := range words {
		if commonTags[word] && !contains(tags, word) {
			tags = append(tags, word)
		}
	}

	if len(tags) == 0 {
		tags = []string{"general", "daily"}
	}

	return tags
}

func (s *Server) analyzeMood(dateStr string) string {
	// Placeholder implementation
	return fmt.Sprintf("Mood analysis for %s: Overall positive trend with minor fluctuations", dateStr)
}

func (s *Server) generateInsights(dateStr string) string {
	// Placeholder implementation
	return fmt.Sprintf("Insights for %s: Productive day with good work-life balance", dateStr)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func main() {
	// Initialize GitHub storage provider
	config := storage.Config{
		StorageType: "github",
		GitHubRepo:  os.Getenv("DAILYLOG_GITHUB_REPO"),
		GitHubToken: os.Getenv("DAILYLOG_GITHUB_TOKEN"),
		GitHubPath:  os.Getenv("DAILYLOG_GITHUB_PATH"),
	}

	// Fallback to default values if env vars not set
	if config.GitHubRepo == "" {
		config.GitHubRepo = "cloudygreybeard/daily-logs" // Replace with actual repo
	}
	if config.GitHubPath == "" {
		config.GitHubPath = "logs"
	}

	storageProvider, err := providers.NewGitHubStorageProvider(config)
	if err != nil {
		log.Fatalf("Failed to create storage provider: %v", err)
	}

	// Verify storage is accessible
	if err := storageProvider.HealthCheck(); err != nil {
		log.Fatalf("Storage health check failed: %v", err)
	}

	// Create our server instance
	dailyLogServer := &Server{storage: storageProvider}

	// Create MCP server with our implementation info
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "dailylog",
		Version: version,
	}, nil)

	// Add daily log tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "dailylog_entry",
		Description: "Create a new daily log entry for activities, moods, notes, or summaries",
	}, dailyLogServer.LogEntry)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dailylog_get_entries",
		Description: "Get log entries for a specific date or date range",
	}, dailyLogServer.GetEntries)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dailylog_search",
		Description: "Search through log entries by text, tags, mood, or other criteria",
	}, dailyLogServer.SearchLogs)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dailylog_summarize",
		Description: "Generate summaries for daily, weekly, monthly, or custom periods",
	}, dailyLogServer.SummarizePeriod)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "dailylog_ai_assist",
		Description: "AI assistance for wording improvements, tag suggestions, mood analysis, and insights",
	}, dailyLogServer.AIAssist)

	// Set up logging to stderr to avoid JSON-RPC interference
	log.SetOutput(os.Stderr)
	log.Println("Starting DailyLog MCP server...")

	// Run the server over stdin/stdout until client disconnects
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal("Server failed:", err)
	}
}
