package log

import "time"

type LogEntry struct {
	Timestamp   time.Time    `json:"timestamp"`
	Level       string       `json:"level"`
	Service     ServiceInfo  `json:"service"`
	Environment string       `json:"environment"`
	Message     string       `json:"message"`
	Context     ContextInfo  `json:"context"`
	HTTP        *HTTPInfo    `json:"http,omitempty"`
	Error       *ErrorInfo   `json:"error,omitempty"`
	Metadata    MetadataInfo `json:"metadata"`
}

type ServiceInfo struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	InstanceID string `json:"instance_id"`
}

type ContextInfo struct {
	UserID    string `json:"user_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

type HTTPInfo struct {
	Method         string `json:"method"`
	URL            string `json:"url"`
	StatusCode     int    `json:"status_code"`
	ResponseTimeMs int    `json:"response_time_ms"`
}

type ErrorInfo struct {
	Code       string `json:"code,omitempty"`
	Message    string `json:"message,omitempty"`
	StackTrace string `json:"stack_trace,omitempty"`
}

type MetadataInfo struct {
	TraceID      string `json:"trace_id,omitempty"`
	SpanID       string `json:"span_id,omitempty"`
	ParentSpanID string `json:"parent_span_id,omitempty"`
	HostIP       string `json:"host_ip,omitempty"`
	ContainerID  string `json:"container_id,omitempty"`
}
