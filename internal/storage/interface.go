package storage

import (
	"time"
)

// DailyLogStorage defines the interface for daily log storage operations
type DailyLogStorage interface {
	// Day operations
	GetDay(date time.Time) (*DayLog, error)
	SaveDay(dayLog *DayLog) error
	DeleteDay(date time.Time) error

	// Entry operations
	CreateEntry(req CreateLogEntryRequest) (*DailyLogEntry, error)
	UpdateEntry(req UpdateLogEntryRequest) (*DailyLogEntry, error)
	DeleteEntry(id string, date time.Time) error
	GetEntry(id string, date time.Time) (*DailyLogEntry, error)

	// Search and retrieval
	SearchLogs(req LogSearchRequest) (*LogSearchResponse, error)
	GetDateRange(start, end time.Time) ([]DayLog, error)
	GetWeek(date time.Time) (*WeeklyLog, error)
	GetMonth(year int, month int) (*MonthlyLog, error)

	// Summary operations
	GenerateSummary(req SummaryRequest) (*SummaryResponse, error)
	SaveSummary(summary *SummaryResponse, targetType string, date time.Time) error

	// Utility operations
	ListDays(start, end time.Time) ([]time.Time, error)
	GetStats(start, end time.Time) (map[string]any, error)
	Backup() error
	HealthCheck() error
}

// BackupStorage defines the interface for backup operations
type BackupStorage interface {
	BackupDay(date time.Time, data []byte) error
	RestoreDay(date time.Time) ([]byte, error)
	ListBackups() ([]time.Time, error)
	DeleteBackup(date time.Time) error
}

// AIProvider defines the interface for AI-powered features
type AIProvider interface {
	GenerateSummary(entries []DailyLogEntry, prompt string) (string, error)
	SuggestTags(description string) ([]string, error)
	AnalyzeStatus(entries []DailyLogEntry) (map[string]any, error)
	GenerateInsights(dayLogs []DayLog) (string, error)
	ImproveWording(text string) (string, error)
}

// Config represents the configuration for the daily log storage
type Config struct {
	StorageType     string `json:"storage_type"` // "github", "local", "cloud"
	GitHubRepo      string `json:"github_repo"`  // "username/repo"
	GitHubToken     string `json:"github_token"` // Personal access token
	GitHubPath      string `json:"github_path"`  // Path within repo
	LocalPath       string `json:"local_path"`   // Local storage path
	BackupEnabled   bool   `json:"backup_enabled"`
	BackupFrequency string `json:"backup_frequency"` // "daily", "weekly"
	AIEnabled       bool   `json:"ai_enabled"`
	AIProvider      string `json:"ai_provider"` // "openai", "anthropic"
	AIAPIKey        string `json:"ai_api_key"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// StorageError represents a storage-related error
type StorageError struct {
	Operation string `json:"operation"`
	Message   string `json:"message"`
	Cause     error  `json:"cause,omitempty"`
}

func (e StorageError) Error() string {
	if e.Cause != nil {
		return e.Operation + ": " + e.Message + " (" + e.Cause.Error() + ")"
	}
	return e.Operation + ": " + e.Message
}

// NotFoundError represents a not found error
type NotFoundError struct {
	Resource string `json:"resource"`
	ID       string `json:"id"`
}

func (e NotFoundError) Error() string {
	return e.Resource + " not found: " + e.ID
}
