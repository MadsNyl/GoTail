package html

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"gotail/models"
	"gotail/ui"
)

func (h *HTMLHandler) HandleLogStatsPage(w http.ResponseWriter, r *http.Request) {
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))

	// If year or month is not provided, use the current date
	if year == 0 || month == 0 {
		now := time.Now()
		year = now.Year()
		month = int(now.Month())
	}

	if year < 2000 || month < 1 || month > 12 {
		http.Error(w, "Invalid year or month", http.StatusBadRequest)
		return
	}

	totalLogs, err := h.Store.CountLogsByMonth(year, month)
	if err != nil {
		http.Error(w, "Failed to fetch total logs", http.StatusInternalServerError)
		return
	}

	severityCounts, err := h.Store.CountLogsBySeverity(year, month)
	if err != nil {
		http.Error(w, "Failed to fetch severity counts", http.StatusInternalServerError)
		return
	}

	dailyCounts, err := h.Store.CountLogsPerDay(year, month)
	if err != nil {
		http.Error(w, "Failed to fetch daily counts", http.StatusInternalServerError)
		return
	}

	// Create a slice and fill missing days
	var sortedDailyCounts []models.DailyCount
	for day := 1; day <= 31; day++ {
		count := dailyCounts[day] // defaults to 0 if not present
		sortedDailyCounts = append(sortedDailyCounts, models.DailyCount{
			Day:   day,
			Count: count,
		})
	}

	// Remove extra days for shorter months
	sortedDailyCounts = sortedDailyCounts[:daysInMonth(year, time.Month(month))]

	serviceCounts, err := h.Store.CountLogsByService(year, month)
	if err != nil {
		http.Error(w, "Failed to fetch service counts", http.StatusInternalServerError)
		return
	}

	attributeCounts, err := h.Store.CountLogsByAttribute(year, month)
	if err != nil {
		http.Error(w, "Failed to fetch attribute counts", http.StatusInternalServerError)
		return
	}

	now := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	prev := now.AddDate(0, -1, 0)
	next := now.AddDate(0, 1, 0)

	rawDailyCounts, err := json.Marshal(sortedDailyCounts)
    if err != nil {
        http.Error(w, "failed to marshal daily counts", http.StatusInternalServerError)
        return
    }

	w.Header().Set("Content-Type", "text/html")
	ui.StatsView(struct {
		Year            int
		Month           int
		TotalLogs       int
		SeverityCounts  map[string]int
		DailyCounts     template.JS
		ServiceCounts   map[string]int
		AttributeCounts map[string]int
		CurrentUrl   	string
		PrevMonth       int
		PrevYear        int
		NextMonth       int
		NextYear        int
	}{
		Year:            year,
		Month:           month,
		TotalLogs:       totalLogs,
		SeverityCounts:  severityCounts,
		DailyCounts:     template.JS(rawDailyCounts),
		ServiceCounts:   serviceCounts,
		AttributeCounts: attributeCounts,
		CurrentUrl:      r.URL.Path,
		PrevMonth: 		 int(prev.Month()),
		PrevYear: 		 prev.Year(),
		NextMonth: 		 int(next.Month()),
		NextYear: 		 next.Year(),
	}).Render(r.Context(), w)
	
}

func daysInMonth(year int, month time.Month) int {
	// Get number of days in the given month
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}