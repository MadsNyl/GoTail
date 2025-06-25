package models

import (
	"time"
)

type Log struct {
	ID                string    `json:"id"`
	Timestamp         time.Time `json:"timestamp"`
	SeverityText      string    `json:"severity_text"`
	SeverityNumber    int       `json:"severity_number"`
	Body              string    `json:"body"`
	ServiceName       *string   `json:"service_name,omitempty"`
	ServiceVersion    *string   `json:"service_version,omitempty"`
	ServiceInstanceID *string   `json:"service_instance_id,omitempty"`
	HostName          *string   `json:"host_name,omitempty"`
	ScopeName         *string   `json:"scope_name,omitempty"`
	ScopeVersion      *string   `json:"scope_version,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type LogEntry struct {
    ID                string            `json:"id"`
    Timestamp         time.Time         `json:"timestamp"`
    SeverityText      string            `json:"severity_text"`
    SeverityNumber    int               `json:"severity_number"`
    Body              string            `json:"body"`
    ServiceName       *string           `json:"service_name,omitempty"`
    ServiceVersion    *string           `json:"service_version,omitempty"`
    ServiceInstanceID *string           `json:"service_instance_id,omitempty"`
    HostName          *string           `json:"host_name,omitempty"`
    ScopeName         *string           `json:"scope_name,omitempty"`
    ScopeVersion      *string           `json:"scope_version,omitempty"`
    CreatedAt         *time.Time        `json:"created_at"`
    Attributes        map[string]any    `json:"attributes,omitempty"`
};
