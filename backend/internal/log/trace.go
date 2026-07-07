package log

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type traceIDKey struct{}

// NewTraceID generates a 12-char hex trace id using crypto/rand.
// Returns "000000000000" (12 zeros) if rand fails (extremely unlikely).
func NewTraceID() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "000000000000"
	}
	return hex.EncodeToString(b)
}

// WithTraceID returns a new context carrying the trace id.
func WithTraceID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, traceIDKey{}, id)
}

// TraceID extracts the trace id from ctx, or returns "" when none is set.
func TraceID(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey{}).(string); ok {
		return v
	}
	return ""
}
