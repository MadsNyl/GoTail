package html

import (
	"log"
	"net/http"
	"strconv"

	"gotail/db"
	"gotail/ui"
    "gotail/models"
)

type HTMLHandler struct {
	Store db.LogStore
}

func (h *HTMLHandler) HandleLogsPage(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    page, _ := strconv.Atoi(q.Get("page"))
    if page < 1 { page = 1 }

    limit, _ := strconv.Atoi(q.Get("limit"))
    if limit < 1 || limit > 100 { limit = 20 }

    severity := q.Get("severity")
	attrKey := q.Get("attr_key")
	attrValue := q.Get("attr_value")
    service := q.Get("service")

    logs, total, err := h.Store.GetLogsFiltered(page, limit, severity, attrKey, attrValue, service)
    if err != nil {
        // Log the error for debugging purposes
        log.Printf("Error fetching logs: %v", err)
        http.Error(w, "Failed to fetch logs", http.StatusInternalServerError)
        return
    }

	attrKeys, err := h.Store.GetAttributeKeys()
	if err != nil {
		http.Error(w, "Failed to fetch attribute keys", http.StatusInternalServerError)
		return
	}

    totalLogs, err := h.Store.GetTotalLogs()
    if err != nil {
        http.Error(w, "Failed to fetch total logs", http.StatusInternalServerError)
        return
    }

    services, err := h.Store.GetServices()
    if err != nil {
        http.Error(w, "Failed to fetch services", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/html")
    ui.LogsView(struct {
        Logs     []models.LogEntry
        Page     int
        Limit    int
        Total    int
        Severity string
		AttrKeys []string
		AttrValue	string
		AttrKey	string
        CurrentUrl string
        TotalLogs int
        Services []string
        Service string
    }{
        Logs:     logs,
        Page:     page,
        Limit:    limit,
        Total:    total,
        Severity: severity,
		AttrKeys: attrKeys,
		AttrValue: attrValue,
		AttrKey: attrKey,
        CurrentUrl: r.URL.Path,
        TotalLogs: totalLogs,
        Services: services,
        Service: service,
    }).Render(r.Context(), w)
}

