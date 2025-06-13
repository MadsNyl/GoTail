package logging

import (
	"encoding/json"
	"net/http"
)

func (h *LogHandler) HandleGetLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.Store.GetLatestLogs(10)
	if err != nil {
		http.Error(w, "Failed to fetch logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}