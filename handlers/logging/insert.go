package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"gotail/db"
	"gotail/models"
)

const (
	maxBodyLen      = 65536 // 64KB
	maxStringLen    = 256
	maxAttributes   = 50
	maxAttrKeyLen   = 256
	maxAttrValueLen = 4096 // 4KB
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
		writeJSONError(w, "Request body too large", http.StatusRequestEntityTooLarge)
		return
	}
	var logEntry models.LogEntry
	if err := json.Unmarshal(body, &logEntry); err != nil {
		writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if msg := validateLogEntry(logEntry); msg != "" {
		writeJSONError(w, msg, http.StatusBadRequest)
		return
	}

	logEntry.ID = uuid.New().String()

	if logEntry.Timestamp.IsZero() {
		logEntry.Timestamp = time.Now().UTC()
	}

	err = h.Store.InsertLog(logEntry)
	if err != nil {
		log.Printf("Failed to insert log: %v", err)
		writeJSONError(w, "Failed to insert log", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func validateLogEntry(e models.LogEntry) string {
	if len(e.Body) > maxBodyLen {
		return fmt.Sprintf("body exceeds maximum length of %d bytes", maxBodyLen)
	}
	if len(e.SeverityText) > maxStringLen {
		return fmt.Sprintf("severity_text exceeds maximum length of %d chars", maxStringLen)
	}
	if e.ServiceName != nil && len(*e.ServiceName) > maxStringLen {
		return fmt.Sprintf("service_name exceeds maximum length of %d chars", maxStringLen)
	}
	if e.HostName != nil && len(*e.HostName) > maxStringLen {
		return fmt.Sprintf("host_name exceeds maximum length of %d chars", maxStringLen)
	}
	if e.ScopeName != nil && len(*e.ScopeName) > maxStringLen {
		return fmt.Sprintf("scope_name exceeds maximum length of %d chars", maxStringLen)
	}
	if len(e.Attributes) > maxAttributes {
		return fmt.Sprintf("attributes exceeds maximum count of %d", maxAttributes)
	}
	for k, v := range e.Attributes {
		if len(k) > maxAttrKeyLen {
			return fmt.Sprintf("attribute key %q exceeds maximum length of %d chars", k, maxAttrKeyLen)
		}
		if str, ok := v.(string); ok && len(str) > maxAttrValueLen {
			return fmt.Sprintf("attribute value for key %q exceeds maximum length of %d bytes", k, maxAttrValueLen)
		}
	}
	return ""
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
