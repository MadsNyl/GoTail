package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"gotail/db"
	"gotail/models"
)

type APIHandler struct {
	Store db.LogStore
}

func (h *APIHandler) HandleLogsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	severity := q.Get("severity")
	attrKey := q.Get("attr_key")
	attrValue := q.Get("attr_value")
	service := q.Get("service")

	logs, total, err := h.Store.GetLogsFiltered(page, limit, severity, attrKey, attrValue, service)
	if err != nil {
		log.Printf("Error fetching logs: %v", err)
		writeError(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}

	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	response := models.LogsAPIResponse{
		Logs:       logs,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	writeJSON(w, response)
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func writeError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.APIError{
		Error: message,
		Code:  code,
	})
}
