package providers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"

	"dailylog/internal/storage"
)

// GitHubStorageProvider implements DailyLogStorage using GitHub as the backend
type GitHubStorageProvider struct {
	client   *github.Client
	ctx      context.Context
	repo     string
	owner    string
	basePath string
	token    string
}

// NewGitHubStorageProvider creates a new GitHub storage provider
func NewGitHubStorageProvider(config storage.Config) (*GitHubStorageProvider, error) {
	if config.GitHubToken == "" {
		return nil, fmt.Errorf("GitHub token is required")
	}
	if config.GitHubRepo == "" {
		return nil, fmt.Errorf("GitHub repository is required")
	}

	// Parse owner/repo from the repo string
	parts := strings.Split(config.GitHubRepo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("GitHub repo must be in format 'owner/repo'")
	}
	owner, repo := parts[0], parts[1]

	// Create OAuth2 token source
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.GitHubToken})
	tc := oauth2.NewClient(context.Background(), ts)

	// Create GitHub client
	client := github.NewClient(tc)

	basePath := config.GitHubPath
	if basePath == "" {
		basePath = "daily-logs"
	}

	return &GitHubStorageProvider{
		client:   client,
		ctx:      context.Background(),
		repo:     repo,
		owner:    owner,
		basePath: basePath,
		token:    config.GitHubToken,
	}, nil
}

// GetDay retrieves a day's log from GitHub
func (g *GitHubStorageProvider) GetDay(date time.Time) (*storage.DayLog, error) {
	filePath := g.getDayFilePath(date)

	fileContent, _, _, err := g.client.Repositories.GetContents(
		g.ctx, g.owner, g.repo, filePath, nil,
	)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			// Create new day log if it doesn't exist
			dayLog := &storage.DayLog{
				Date:         date,
				Entries:      []storage.DailyLogEntry{},
				TotalEntries: 0,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}
			return dayLog, nil
		}
		return nil, storage.StorageError{
			Operation: "GetDay",
			Message:   fmt.Sprintf("failed to get day %s", date.Format("2006-01-02")),
			Cause:     err,
		}
	}

	// Decode the content
	content, err := base64.StdEncoding.DecodeString(*fileContent.Content)
	if err != nil {
		return nil, storage.StorageError{
			Operation: "GetDay",
			Message:   "failed to decode file content",
			Cause:     err,
		}
	}

	var dayLog storage.DayLog
	if err := json.Unmarshal(content, &dayLog); err != nil {
		return nil, storage.StorageError{
			Operation: "GetDay",
			Message:   "failed to parse day log JSON",
			Cause:     err,
		}
	}

	return &dayLog, nil
}

// SaveDay saves a day's log to GitHub
func (g *GitHubStorageProvider) SaveDay(dayLog *storage.DayLog) error {
	filePath := g.getDayFilePath(dayLog.Date)

	// Convert to JSON
	content, err := dayLog.ToJSON()
	if err != nil {
		return storage.StorageError{
			Operation: "SaveDay",
			Message:   "failed to serialize day log",
			Cause:     err,
		}
	}

	// Check if file exists to get SHA
	var sha *string
	existingFile, _, _, err := g.client.Repositories.GetContents(
		g.ctx, g.owner, g.repo, filePath, nil,
	)
	if err == nil && existingFile != nil {
		sha = existingFile.SHA
	}

	// Create commit message
	commitMessage := fmt.Sprintf("Update daily log for %s", dayLog.GetDateString())
	if sha == nil {
		commitMessage = fmt.Sprintf("Create daily log for %s", dayLog.GetDateString())
	}

	// Create or update the file
	_, _, err = g.client.Repositories.CreateFile(
		g.ctx, g.owner, g.repo, filePath,
		&github.RepositoryContentFileOptions{
			Message: &commitMessage,
			Content: content,
			SHA:     sha,
		},
	)

	if err != nil {
		return storage.StorageError{
			Operation: "SaveDay",
			Message:   fmt.Sprintf("failed to save day %s", dayLog.GetDateString()),
			Cause:     err,
		}
	}

	return nil
}

// DeleteDay deletes a day's log from GitHub
func (g *GitHubStorageProvider) DeleteDay(date time.Time) error {
	filePath := g.getDayFilePath(date)

	// Get the file to obtain its SHA
	fileContent, _, _, err := g.client.Repositories.GetContents(
		g.ctx, g.owner, g.repo, filePath, nil,
	)
	if err != nil {
		return storage.NotFoundError{
			Resource: "day log",
			ID:       date.Format("2006-01-02"),
		}
	}

	// Delete the file
	commitMessage := fmt.Sprintf("Delete daily log for %s", date.Format("2006-01-02"))
	_, _, err = g.client.Repositories.DeleteFile(
		g.ctx, g.owner, g.repo, filePath,
		&github.RepositoryContentFileOptions{
			Message: &commitMessage,
			SHA:     fileContent.SHA,
		},
	)

	if err != nil {
		return storage.StorageError{
			Operation: "DeleteDay",
			Message:   fmt.Sprintf("failed to delete day %s", date.Format("2006-01-02")),
			Cause:     err,
		}
	}

	return nil
}

// CreateEntry creates a new log entry for a specific day
func (g *GitHubStorageProvider) CreateEntry(req storage.CreateLogEntryRequest) (*storage.DailyLogEntry, error) {
	// Get the day log
	dayLog, err := g.GetDay(req.Date)
	if err != nil {
		return nil, err
	}

	// Create new entry with ID
	entry := storage.DailyLogEntry{
		ID:          g.generateEntryID(),
		Timestamp:   req.Date,
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Tags:        req.Tags,
		Location:    req.Location,
		Metadata:    req.Metadata,
	}

	if req.Status != nil {
		entry.Status = *req.Status
	}
	if req.Priority != nil {
		entry.Priority = *req.Priority
	}
	if req.Duration != nil {
		entry.Duration = req.Duration
	}

	// Add entry to day log
	dayLog.AddEntry(entry)

	// Save the updated day log
	if err := g.SaveDay(dayLog); err != nil {
		return nil, err
	}

	return &entry, nil
}

// UpdateEntry updates an existing log entry
func (g *GitHubStorageProvider) UpdateEntry(req storage.UpdateLogEntryRequest) (*storage.DailyLogEntry, error) {
	// Find the day that contains this entry
	// For now, we'll need to search recent days or require the date to be provided
	// This is a limitation of the current design - we should include date in the request
	return nil, fmt.Errorf("UpdateEntry not implemented - requires date in request")
}

// DeleteEntry deletes a log entry from a specific day
func (g *GitHubStorageProvider) DeleteEntry(id string, date time.Time) error {
	// Get the day log
	dayLog, err := g.GetDay(date)
	if err != nil {
		return err
	}

	// Remove the entry
	if !dayLog.RemoveEntry(id) {
		return storage.NotFoundError{
			Resource: "log entry",
			ID:       id,
		}
	}

	// Save the updated day log
	return g.SaveDay(dayLog)
}

// GetEntry retrieves a specific log entry from a day
func (g *GitHubStorageProvider) GetEntry(id string, date time.Time) (*storage.DailyLogEntry, error) {
	dayLog, err := g.GetDay(date)
	if err != nil {
		return nil, err
	}

	for _, entry := range dayLog.Entries {
		if entry.ID == id {
			return &entry, nil
		}
	}

	return nil, storage.NotFoundError{
		Resource: "log entry",
		ID:       id,
	}
}

// SearchLogs searches through logs based on criteria
func (g *GitHubStorageProvider) SearchLogs(req storage.LogSearchRequest) (*storage.LogSearchResponse, error) {
	// This is a simplified implementation - in reality, we'd need to iterate through files
	// or maintain an index for efficient searching
	response := &storage.LogSearchResponse{
		Entries:     []storage.DailyLogEntry{},
		TotalCount:  0,
		SearchQuery: req,
	}

	// For now, search within a reasonable date range
	startDate := time.Now().AddDate(0, -3, 0) // Last 3 months
	endDate := time.Now()

	if req.DateStart != nil {
		startDate = *req.DateStart
	}
	if req.DateEnd != nil {
		endDate = *req.DateEnd
	}

	// Iterate through date range
	for d := startDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
		dayLog, err := g.GetDay(d)
		if err != nil {
			continue // Skip days that don't exist or have errors
		}

		// Filter entries based on search criteria
		for _, entry := range dayLog.Entries {
			if g.matchesSearchCriteria(entry, req) {
				response.Entries = append(response.Entries, entry)
				response.TotalCount++

				// Respect limit
				if req.Limit > 0 && response.TotalCount >= req.Limit {
					return response, nil
				}
			}
		}
	}

	return response, nil
}

// GetDateRange retrieves all day logs within a date range
func (g *GitHubStorageProvider) GetDateRange(start, end time.Time) ([]storage.DayLog, error) {
	var dayLogs []storage.DayLog

	for d := start; d.Before(end) || d.Equal(end); d = d.AddDate(0, 0, 1) {
		dayLog, err := g.GetDay(d)
		if err != nil {
			continue // Skip days that don't exist
		}
		if len(dayLog.Entries) > 0 {
			dayLogs = append(dayLogs, *dayLog)
		}
	}

	return dayLogs, nil
}

// GetWeek retrieves a week's worth of logs
func (g *GitHubStorageProvider) GetWeek(date time.Time) (*storage.WeeklyLog, error) {
	// Calculate week start (Monday)
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}
	weekStart := date.AddDate(0, 0, -(weekday - 1))
	weekEnd := weekStart.AddDate(0, 0, 6)

	days, err := g.GetDateRange(weekStart, weekEnd)
	if err != nil {
		return nil, err
	}

	weeklyLog := &storage.WeeklyLog{
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
		Days:      days,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Calculate total entries
	for _, day := range days {
		weeklyLog.TotalEntries += day.TotalEntries
	}

	return weeklyLog, nil
}

// GetMonth retrieves a month's worth of logs
func (g *GitHubStorageProvider) GetMonth(year int, month int) (*storage.MonthlyLog, error) {
	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, -1)

	days, err := g.GetDateRange(monthStart, monthEnd)
	if err != nil {
		return nil, err
	}

	monthlyLog := &storage.MonthlyLog{
		Month:     fmt.Sprintf("%04d-%02d", year, month),
		Year:      year,
		Days:      days,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Calculate total entries
	for _, day := range days {
		monthlyLog.TotalEntries += day.TotalEntries
	}

	return monthlyLog, nil
}

// GenerateSummary generates a summary for the given request
func (g *GitHubStorageProvider) GenerateSummary(req storage.SummaryRequest) (*storage.SummaryResponse, error) {
	// Basic implementation - this would integrate with AI in a real implementation
	var summary string
	var stats map[string]any

	switch req.Type {
	case "day":
		dayLog, err := g.GetDay(req.Date)
		if err != nil {
			return nil, err
		}
		summary = g.generateDaySummary(dayLog)
		stats = map[string]any{
			"total_entries":  dayLog.TotalEntries,
			"status_average": dayLog.StatusAverage,
		}

	case "week":
		weekLog, err := g.GetWeek(req.Date)
		if err != nil {
			return nil, err
		}
		summary = g.generateWeekSummary(weekLog)
		stats = map[string]any{
			"total_entries": weekLog.TotalEntries,
			"total_days":    len(weekLog.Days),
		}

	case "month":
		monthLog, err := g.GetMonth(req.Date.Year(), int(req.Date.Month()))
		if err != nil {
			return nil, err
		}
		summary = g.generateMonthSummary(monthLog)
		stats = map[string]any{
			"total_entries": monthLog.TotalEntries,
			"total_days":    len(monthLog.Days),
		}
	}

	return &storage.SummaryResponse{
		Summary:   summary,
		Type:      req.Type,
		Period:    req.Date.Format("2006-01-02"),
		Stats:     stats,
		CreatedAt: time.Now(),
	}, nil
}

// SaveSummary saves a summary to the appropriate location
func (g *GitHubStorageProvider) SaveSummary(summary *storage.SummaryResponse, targetType string, date time.Time) error {
	// Save summary as metadata in the day/week/month file
	switch targetType {
	case "day":
		dayLog, err := g.GetDay(date)
		if err != nil {
			return err
		}
		dayLog.DaySummary = summary.Summary
		return g.SaveDay(dayLog)
	}

	return nil
}

// ListDays lists all available days within a date range
func (g *GitHubStorageProvider) ListDays(start, end time.Time) ([]time.Time, error) {
	var dates []time.Time

	// List files in the repository to find existing days
	// This is a simplified implementation
	for d := start; d.Before(end) || d.Equal(end); d = d.AddDate(0, 0, 1) {
		_, _, _, err := g.client.Repositories.GetContents(
			g.ctx, g.owner, g.repo, g.getDayFilePath(d), nil,
		)
		if err == nil {
			dates = append(dates, d)
		}
	}

	return dates, nil
}

// GetStats returns statistics for a date range
func (g *GitHubStorageProvider) GetStats(start, end time.Time) (map[string]any, error) {
	days, err := g.GetDateRange(start, end)
	if err != nil {
		return nil, err
	}

	totalEntries := 0
	totalDays := len(days)
	statusSum := 0.0
	statusCount := 0

	for _, day := range days {
		totalEntries += day.TotalEntries
		if day.StatusAverage > 0 {
			statusSum += day.StatusAverage
			statusCount++
		}
	}

	avgStatus := 0.0
	if statusCount > 0 {
		avgStatus = statusSum / float64(statusCount)
	}

	return map[string]any{
		"total_entries":   totalEntries,
		"total_days":      totalDays,
		"average_status":  avgStatus,
		"entries_per_day": float64(totalEntries) / float64(totalDays),
	}, nil
}

// Backup creates a backup of all data
func (g *GitHubStorageProvider) Backup() error {
	// GitHub itself is the backup - this could create a separate backup repo
	return nil
}

// HealthCheck verifies the storage is accessible
func (g *GitHubStorageProvider) HealthCheck() error {
	// Try to access the repository
	_, _, err := g.client.Repositories.Get(g.ctx, g.owner, g.repo)
	if err != nil {
		return storage.StorageError{
			Operation: "HealthCheck",
			Message:   "failed to access GitHub repository",
			Cause:     err,
		}
	}
	return nil
}

// Helper methods

func (g *GitHubStorageProvider) getDayFilePath(date time.Time) string {
	return path.Join(g.basePath, date.Format("2006"), date.Format("01"), date.Format("2006-01-02.json"))
}

func (g *GitHubStorageProvider) generateEntryID() string {
	return fmt.Sprintf("entry_%d", time.Now().UnixNano())
}

func (g *GitHubStorageProvider) matchesSearchCriteria(entry storage.DailyLogEntry, req storage.LogSearchRequest) bool {
	// Type filter
	if req.Type != "" && entry.Type != req.Type {
		return false
	}

	// Status range filter
	if req.StatusMin != nil && entry.Status < *req.StatusMin {
		return false
	}
	if req.StatusMax != nil && entry.Status > *req.StatusMax {
		return false
	}

	// Text search in title and description
	if req.SearchText != "" {
		searchText := strings.ToLower(req.SearchText)
		if !strings.Contains(strings.ToLower(entry.Title), searchText) &&
			!strings.Contains(strings.ToLower(entry.Description), searchText) {
			return false
		}
	}

	// Tag filter
	if len(req.Tags) > 0 {
		hasTag := false
		for _, reqTag := range req.Tags {
			for _, entryTag := range entry.Tags {
				if entryTag == reqTag {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}
		if !hasTag {
			return false
		}
	}

	return true
}

func (g *GitHubStorageProvider) generateDaySummary(dayLog *storage.DayLog) string {
	if len(dayLog.Entries) == 0 {
		return "No activities recorded for this day."
	}

	return fmt.Sprintf("Day had %d activities with an average status of %.1f",
		dayLog.TotalEntries, dayLog.StatusAverage)
}

func (g *GitHubStorageProvider) generateWeekSummary(weekLog *storage.WeeklyLog) string {
	return fmt.Sprintf("Week had %d total activities across %d days",
		weekLog.TotalEntries, len(weekLog.Days))
}

func (g *GitHubStorageProvider) generateMonthSummary(monthLog *storage.MonthlyLog) string {
	return fmt.Sprintf("Month had %d total activities across %d days",
		monthLog.TotalEntries, len(monthLog.Days))
}
