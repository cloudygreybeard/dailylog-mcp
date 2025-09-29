package storage

import (
	"encoding/json"
	"time"
)

// DailyLogEntry represents a single activity or entry in a day
type DailyLogEntry struct {
	ID          string            `json:"id"`
	Timestamp   time.Time         `json:"timestamp"`
	Type        string            `json:"type"` // "activity", "mood", "note", "summary"
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags,omitempty"`
	Mood        int               `json:"mood,omitempty"`     // 1-10 scale
	Priority    int               `json:"priority,omitempty"` // 1-5 scale
	Duration    *int              `json:"duration,omitempty"` // minutes
	Location    string            `json:"location,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// DayLog represents all activities and entries for a single day
type DayLog struct {
	Date         time.Time       `json:"date"`
	Entries      []DailyLogEntry `json:"entries"`
	DaySummary   string          `json:"day_summary,omitempty"`
	MoodAverage  float64         `json:"mood_average,omitempty"`
	TotalEntries int             `json:"total_entries"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	Metadata     map[string]any  `json:"metadata,omitempty"`
}

// WeeklyLog represents a week's worth of daily logs
type WeeklyLog struct {
	WeekStart    time.Time `json:"week_start"`
	WeekEnd      time.Time `json:"week_end"`
	Days         []DayLog  `json:"days"`
	WeekSummary  string    `json:"week_summary,omitempty"`
	Highlights   []string  `json:"highlights,omitempty"`
	Challenges   []string  `json:"challenges,omitempty"`
	TotalEntries int       `json:"total_entries"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// MonthlyLog represents a month's worth of daily logs
type MonthlyLog struct {
	Month        string    `json:"month"` // "2025-09"
	Year         int       `json:"year"`
	Days         []DayLog  `json:"days"`
	MonthSummary string    `json:"month_summary,omitempty"`
	Goals        []string  `json:"goals,omitempty"`
	Achievements []string  `json:"achievements,omitempty"`
	TotalEntries int       `json:"total_entries"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// LogSearchRequest represents parameters for searching logs
type LogSearchRequest struct {
	DateStart  *time.Time        `json:"date_start,omitempty"`
	DateEnd    *time.Time        `json:"date_end,omitempty"`
	Type       string            `json:"type,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
	MoodMin    *int              `json:"mood_min,omitempty"`
	MoodMax    *int              `json:"mood_max,omitempty"`
	SearchText string            `json:"search_text,omitempty"`
	Limit      int               `json:"limit,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// LogSearchResponse represents the result of a log search
type LogSearchResponse struct {
	Entries     []DailyLogEntry  `json:"entries"`
	Days        []DayLog         `json:"days,omitempty"`
	TotalCount  int              `json:"total_count"`
	SearchQuery LogSearchRequest `json:"search_query"`
}

// CreateLogEntryRequest represents a request to create a new log entry
type CreateLogEntryRequest struct {
	Date        time.Time         `json:"date"`
	Type        string            `json:"type"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags,omitempty"`
	Mood        *int              `json:"mood,omitempty"`
	Priority    *int              `json:"priority,omitempty"`
	Duration    *int              `json:"duration,omitempty"`
	Location    string            `json:"location,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// UpdateLogEntryRequest represents a request to update an existing log entry
type UpdateLogEntryRequest struct {
	ID          string            `json:"id"`
	Type        string            `json:"type,omitempty"`
	Title       string            `json:"title,omitempty"`
	Description string            `json:"description,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Mood        *int              `json:"mood,omitempty"`
	Priority    *int              `json:"priority,omitempty"`
	Duration    *int              `json:"duration,omitempty"`
	Location    string            `json:"location,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// SummaryRequest represents a request to generate a summary
type SummaryRequest struct {
	Type      string     `json:"type"` // "day", "week", "month"
	Date      time.Time  `json:"date"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	UseAI     bool       `json:"use_ai"`
	Prompt    string     `json:"prompt,omitempty"`
}

// SummaryResponse represents the result of a summary generation
type SummaryResponse struct {
	Summary   string            `json:"summary"`
	Type      string            `json:"type"`
	Period    string            `json:"period"`
	Stats     map[string]any    `json:"stats"`
	CreatedAt time.Time         `json:"created_at"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Utility methods for DayLog

// AddEntry adds a new entry to the day log
func (d *DayLog) AddEntry(entry DailyLogEntry) {
	d.Entries = append(d.Entries, entry)
	d.TotalEntries = len(d.Entries)
	d.UpdatedAt = time.Now()
	d.calculateMoodAverage()
}

// UpdateEntry updates an existing entry in the day log
func (d *DayLog) UpdateEntry(id string, updatedEntry DailyLogEntry) bool {
	for i, entry := range d.Entries {
		if entry.ID == id {
			d.Entries[i] = updatedEntry
			d.UpdatedAt = time.Now()
			d.calculateMoodAverage()
			return true
		}
	}
	return false
}

// RemoveEntry removes an entry from the day log
func (d *DayLog) RemoveEntry(id string) bool {
	for i, entry := range d.Entries {
		if entry.ID == id {
			d.Entries = append(d.Entries[:i], d.Entries[i+1:]...)
			d.TotalEntries = len(d.Entries)
			d.UpdatedAt = time.Now()
			d.calculateMoodAverage()
			return true
		}
	}
	return false
}

// calculateMoodAverage calculates the average mood for the day
func (d *DayLog) calculateMoodAverage() {
	var total float64
	var count int

	for _, entry := range d.Entries {
		if entry.Mood > 0 {
			total += float64(entry.Mood)
			count++
		}
	}

	if count > 0 {
		d.MoodAverage = total / float64(count)
	} else {
		d.MoodAverage = 0
	}
}

// GetEntriesByType returns all entries of a specific type
func (d *DayLog) GetEntriesByType(entryType string) []DailyLogEntry {
	var entries []DailyLogEntry
	for _, entry := range d.Entries {
		if entry.Type == entryType {
			entries = append(entries, entry)
		}
	}
	return entries
}

// GetEntriesByTag returns all entries containing a specific tag
func (d *DayLog) GetEntriesByTag(tag string) []DailyLogEntry {
	var entries []DailyLogEntry
	for _, entry := range d.Entries {
		for _, entryTag := range entry.Tags {
			if entryTag == tag {
				entries = append(entries, entry)
				break
			}
		}
	}
	return entries
}

// ToJSON converts the DayLog to JSON
func (d *DayLog) ToJSON() ([]byte, error) {
	return json.MarshalIndent(d, "", "  ")
}

// FromJSON creates a DayLog from JSON bytes
func (d *DayLog) FromJSON(data []byte) error {
	return json.Unmarshal(data, d)
}

// GetDateString returns the date as a string in YYYY-MM-DD format
func (d *DayLog) GetDateString() string {
	return d.Date.Format("2006-01-02")
}

// GetFilename returns the expected filename for this day log
func (d *DayLog) GetFilename() string {
	return d.Date.Format("2006-01-02") + ".json"
}
