package httpapi

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRouter(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{name: "healthz", path: "/healthz", wantStatus: http.StatusOK, wantBody: `{"status":"ok"}`},
		{name: "unknown route", path: "/nope", wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			h := NewRouter(slog.New(slog.NewJSONHandler(&logs, nil)))

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && strings.TrimSpace(rec.Body.String()) != tt.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}

			var entry map[string]any
			if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
				t.Fatalf("request log is not JSON: %v (%q)", err, logs.String())
			}
			for _, k := range []string{"method", "path", "status", "duration_ms", "request_id"} {
				if _, ok := entry[k]; !ok {
					t.Errorf("request log missing %q: %v", k, entry)
				}
			}
			if got := int(entry["status"].(float64)); got != tt.wantStatus {
				t.Errorf("logged status = %d, want %d", got, tt.wantStatus)
			}
		})
	}
}

func TestRequestLogger_NeverLogsBodiesOrHeaders(t *testing.T) {
	const bodySentinel = "BODY-SENTINEL-4f1c"
	const headerSentinel = "HEADER-SENTINEL-9a2e"

	var logs bytes.Buffer
	h := NewRouter(slog.New(slog.NewJSONHandler(&logs, nil)))

	req := httptest.NewRequest(http.MethodPost, "/healthz", strings.NewReader(bodySentinel))
	req.Header.Set("Authorization", "Bearer "+headerSentinel)
	req.Header.Set("Cookie", "session="+headerSentinel)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if logs.Len() == 0 {
		t.Fatal("expected a request log line")
	}
	for _, s := range []string{bodySentinel, headerSentinel} {
		if strings.Contains(logs.String(), s) {
			t.Errorf("log contains sentinel %q: %s", s, logs.String())
		}
	}
}

func TestRecoverer_Returns500WithoutLoggingPanicValue(t *testing.T) {
	const secret = "PANIC-SECRET-7d3b"

	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	r := chi.NewRouter()
	r.Use(requestLogger(logger))
	r.Use(recoverer(logger))
	r.Get("/boom", func(http.ResponseWriter, *http.Request) { panic(secret) })

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if !strings.Contains(logs.String(), "panic recovered") {
		t.Errorf("expected panic to be logged: %s", logs.String())
	}
	if strings.Contains(logs.String(), secret) {
		t.Errorf("log contains panic value: %s", logs.String())
	}
}
