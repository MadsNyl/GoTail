package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gotail/db"
)

type LogHandler struct {
	Store db.LogStore
}

func (h *LogHandler) HandleLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST supported", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	var logEntry db.LogEntry
	if err := json.Unmarshal(body, &logEntry); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if logEntry.ID == "" {
		logEntry.ID = uuid.NewString()
	}
	if logEntry.Timestamp.IsZero() {
		logEntry.Timestamp = time.Now().UTC()
	}

	err = h.Store.InsertLog(logEntry)
	if err != nil {
		http.Error(w, "Failed to insert log", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *LogHandler) HandleGetLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.Store.GetRecentLogs(100) // limit to 100 for now
	if err != nil {
		http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}