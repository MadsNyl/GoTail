package logging

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"gotail/db"
	"gotail/models"
)

type LogHandler struct {
	Store db.LogStore
}

func (h *LogHandler) HandleLogInsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST supported", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	var logEntry models.LogEntry
	if err := json.Unmarshal(body, &logEntry); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	logEntry.ID = uuid.New().String()

	if logEntry.Timestamp.IsZero() {
		logEntry.Timestamp = time.Now().UTC()
	}

	err = h.Store.InsertLog(logEntry)
	if err != nil {
		log.Printf("Failed to insert log: %v", err)
		http.Error(w, "Failed to insert log", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}