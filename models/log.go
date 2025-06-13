package models

import (
	"time"
)

type LogEntry struct {
    ID                 string            	`json:"id"`
    Timestamp          time.Time         	`json:"timestamp"`
    ObservedTimestamp  *time.Time        	`json:"observed_timestamp,omitempty"`
    TraceID            string            	`json:"trace_id,omitempty"`
    SpanID             string            	`json:"span_id,omitempty"`
    TraceFlags         *uint8            	`json:"trace_flags,omitempty"`
    SeverityText       string            	`json:"severity_text"`
    SeverityNumber     int               	`json:"severity_number"`
    Body               string            	`json:"body"`
    Resource           map[string]string 	`json:"resource,omitempty"`
    InstrumentationScope map[string]string 	`json:"instrumentation_scope,omitempty"`
    Attributes         map[string]string 	`json:"attributes,omitempty"`
    CreatedAt          time.Time         	`json:"created_at"`
    ServiceName        string            	`json:"service_name,omitempty"`
    ServiceVersion     string            	`json:"service_version,omitempty"`
    ServiceInstanceID  string            	`json:"service_instance_id,omitempty"`
    HostName           string            	`json:"host_name,omitempty"`
    ScopeName          string            	`json:"scope_name,omitempty"`
    ScopeVersion       string            	`json:"scope_version,omitempty"`
}
