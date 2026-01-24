package models

type LogsAPIResponse struct {
	Logs       []LogEntry `json:"logs"`
	Page       int        `json:"page"`
	Limit      int        `json:"limit"`
	Total      int        `json:"total"`
	TotalPages int        `json:"total_pages"`
}

type StatsAPIResponse struct {
	Year            int            `json:"year"`
	Month           int            `json:"month"`
	TotalLogs       int            `json:"total_logs"`
	SeverityCounts  map[string]int `json:"severity_counts"`
	DailyCounts     []DailyCount   `json:"daily_counts"`
	ServiceCounts   map[string]int `json:"service_counts"`
	AttributeCounts map[string]int `json:"attribute_counts"`
}

type APIError struct {
	Error string `json:"error"`
	Code  int    `json:"code,omitempty"`
}
