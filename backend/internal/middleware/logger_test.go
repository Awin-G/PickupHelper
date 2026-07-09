package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// captureLogs swaps the default slog handler with a JSON handler writing to
// a buffer, runs fn, then restores the previous default. Returns the buffer.
func captureLogs(t *testing.T, fn func()) *bytes.Buffer {
	t.Helper()
	prev := slog.Default()
	buf := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	fn()
	return buf
}

func TestLogger_WritesOneLinePerRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	buf := captureLogs(t, func() {
		engine := gin.New()
		engine.Use(TraceID(), Logger())
		engine.GET("/x", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest(http.MethodGet, "/x?foo=bar", nil)
		rr := httptest.NewRecorder()
		engine.ServeHTTP(rr, req)
	})

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 log line, got %d (%q)", len(lines), buf.String())
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("decode log line: %v", err)
	}
	if entry["msg"] != "http" {
		t.Errorf("msg = %v, want 'http'", entry["msg"])
	}
	if entry["method"] != "GET" {
		t.Errorf("method = %v, want GET", entry["method"])
	}
	if entry["path"] != "/x" {
		t.Errorf("path = %v, want /x", entry["path"])
	}
	if entry["query"] != "foo=bar" {
		t.Errorf("query = %v, want foo=bar", entry["query"])
	}
	if entry["status"] != float64(200) {
		t.Errorf("status = %v, want 200", entry["status"])
	}
	if entry["trace_id"] == "" || entry["trace_id"] == "-" {
		t.Errorf("trace_id should be populated, got %v", entry["trace_id"])
	}
}

func TestLogger_5xxLogsAtErrorLevel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	buf := captureLogs(t, func() {
		engine := gin.New()
		// Logger outermost so it observes the final status set by Recovery.
		engine.Use(Logger(), Recovery())
		engine.GET("/boom", func(c *gin.Context) {
			panic("x")
		})

		req := httptest.NewRequest(http.MethodGet, "/boom", nil)
		rr := httptest.NewRecorder()
		engine.ServeHTTP(rr, req)
	})

	// Two log lines: panic-recovered (ERROR) + http (ERROR). Find the http one.
	var httpEntry map[string]any
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if entry["msg"] == "http" {
			httpEntry = entry
			break
		}
	}
	if httpEntry == nil {
		t.Fatalf("no http log line found in:\n%s", buf.String())
	}
	if level, _ := httpEntry["level"].(string); level != "ERROR" {
		t.Errorf("level = %v, want ERROR", level)
	}
	if httpEntry["status"] != float64(500) {
		t.Errorf("status = %v, want 500", httpEntry["status"])
	}
}
