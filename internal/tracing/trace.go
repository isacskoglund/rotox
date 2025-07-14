// Package tracing provides distributed tracing utilities for rotox.
//
// This package implements context-based trace ID propagation and
// automatic injection into structured logs. It enables correlation
// of log entries and operations across the distributed system.
//
// Trace IDs are propagated through contexts and automatically
// added to log entries, making it easier to track requests
// through the hub-probe architecture.
package tracing

import (
	"context"
	"log/slog"
)

// contextKey is used to store trace IDs in context values.
type contextKey string

// traceKey is the context key for storing trace IDs.
const traceKey contextKey = "trace_id"

// WithTraceId adds a trace ID to the provided context.
// The trace ID will be automatically included in log entries
// when using a tracing-enabled logger.
func WithTraceId(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, traceKey, traceId)
}

// GetTraceId extracts the trace ID from the provided context.
// Returns an empty string if no trace ID is present.
func GetTraceId(ctx context.Context) string {
	traceId, ok := ctx.Value(traceKey).(string)
	if !ok {
		return ""
	}
	return traceId
}

// tracingHandler wraps an slog.Handler to automatically inject
// trace IDs into log records when present in the context.
type tracingHandler struct {
	slog.Handler
}

// NewHandler creates a new tracing-enabled log handler that wraps
// the provided handler. It will automatically add trace IDs to
// log records when available in the logging context.
func NewHandler(h slog.Handler) slog.Handler {
	return &tracingHandler{
		Handler: h,
	}
}

func (h *tracingHandler) Handle(ctx context.Context, r slog.Record) error {
	if traceId, ok := ctx.Value(traceKey).(string); ok {
		r.AddAttrs(slog.String(string(traceKey), traceId))
	}
	return h.Handler.Handle(ctx, r)
}
