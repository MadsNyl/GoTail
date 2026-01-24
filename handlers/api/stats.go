package api

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"gotail/models"
)

func (h *APIHandler) HandleStatsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))

	// If year or month is not provided, use the current date
	if year == 0 || month == 0 {
		now := time.Now()
		year = now.Year()
		month = int(now.Month())
	}

	if year < 2000 || month < 1 || month > 12 {
		writeError(w, "Invalid year or month", http.StatusBadRequest)
		return
	}

	totalLogs, err := h.Store.CountLogsByMonth(year, month)
	if err != nil {
		log.Printf("Error fetching total logs: %v", err)
		writeError(w, "Failed to fetch total logs", http.StatusInternalServerError)
		return
	}

	severityCounts, err := h.Store.CountLogsBySeverity(year, month)
	if err != nil {
		log.Printf("Error fetching severity counts: %v", err)
		writeError(w, "Failed to fetch severity counts", http.StatusInternalServerError)
		return
	}

	dailyCountsMap, err := h.Store.CountLogsPerDay(year, month)
	if err != nil {
		log.Printf("Error fetching daily counts: %v", err)
		writeError(w, "Failed to fetch daily counts", http.StatusInternalServerError)
		return
	}

	// Convert map to sorted slice and fill missing days
	daysInMonth := daysInMonth(year, time.Month(month))
	dailyCounts := make([]models.DailyCount, 0, daysInMonth)
	for day := 1; day <= daysInMonth; day++ {
		count := dailyCountsMap[day]
		dailyCounts = append(dailyCounts, models.DailyCount{
			Day:   day,
			Count: count,
		})
	}

	serviceCounts, err := h.Store.CountLogsByService(year, month)
	if err != nil {
		log.Printf("Error fetching service counts: %v", err)
		writeError(w, "Failed to fetch service counts", http.StatusInternalServerError)
		return
	}

	attributeCounts, err := h.Store.CountLogsByAttribute(year, month)
	if err != nil {
		log.Printf("Error fetching attribute counts: %v", err)
		writeError(w, "Failed to fetch attribute counts", http.StatusInternalServerError)
		return
	}

	response := models.StatsAPIResponse{
		Year:            year,
		Month:           month,
		TotalLogs:       totalLogs,
		SeverityCounts:  severityCounts,
		DailyCounts:     dailyCounts,
		ServiceCounts:   serviceCounts,
		AttributeCounts: attributeCounts,
	}

	writeJSON(w, response)
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
