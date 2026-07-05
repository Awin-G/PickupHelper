package log

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTraceID(t *testing.T) {
	id := NewTraceID()
	// 12 hex chars = 6 bytes hex-encoded.
	assert.Len(t, id, 12)
	// All chars must be hex.
	for _, c := range id {
		assert.True(t, strings.ContainsRune("0123456789abcdef", c), "non-hex char %q", c)
	}
}

func TestWithTraceID(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, "", TraceID(ctx))
	ctx = WithTraceID(ctx, "abc123")
	assert.Equal(t, "abc123", TraceID(ctx))
	// Empty id should not set the value.
	ctx2 := WithTraceID(context.Background(), "")
	assert.Equal(t, "", TraceID(ctx2))
}

func TestHandler_InjectsTraceID(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(&buf, "info")
	logger := slog.New(h)
	ctx := WithTraceID(context.Background(), "abc123")
	logger.InfoContext(ctx, "hello", "user", "u1")

	var out map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, "abc123", out["trace_id"])
	assert.Equal(t, "hello", out["msg"])
	assert.Equal(t, "u1", out["user"])
}

func TestHandler_NoTraceID(t *testing.T) {
	var buf bytes.Buffer
	h := NewHandler(&buf, "info")
	logger := slog.New(h)
	logger.InfoContext(context.Background(), "x")

	var out map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &out))
	assert.Equal(t, "-", out["trace_id"])
}
