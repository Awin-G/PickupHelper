package log

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Handler wraps slog.JSONHandler and injects trace_id from context into every record.
type Handler struct {
	inner slog.Handler
}

// NewHandler constructs a Handler writing JSON to w at the given level.
func NewHandler(w io.Writer, level string) *Handler {
	lvl := parseLevel(level)
	inner := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: lvl})
	return &Handler{inner: inner}
}

// Enabled delegates to the wrapped handler.
func (h *Handler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return h.inner.Enabled(ctx, lvl)
}

// Handle injects trace_id (or "-") into the record before delegating.
func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	tid := TraceID(ctx)
	if tid == "" {
		tid = "-"
	}
	r.AddAttrs(slog.String("trace_id", tid))
	return h.inner.Handle(ctx, r)
}

// WithAttrs delegates.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{inner: h.inner.WithAttrs(attrs)}
}

// WithGroup delegates.
func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{inner: h.inner.WithGroup(name)}
}

// Init initializes the global slog default logger writing JSON to stdout.
func Init(level string) {
	slog.SetDefault(slog.New(NewHandler(os.Stdout, level)))
}

// From returns the default logger. The Handler injects trace_id from the
// context passed to *Context log methods (InfoContext, ErrorContext, ...).
// Callers that want trace_id in their log line MUST pass ctx via the
// *Context family of methods.
func From(ctx context.Context) *slog.Logger {
	_ = ctx
	return slog.Default()
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
