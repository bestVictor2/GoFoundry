package Gee

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddleware_Generate(t *testing.T) {
	engine := New()
	engine.Use(RequestID())
	engine.GET("/ping", func(c *Context) {
		c.String(http.StatusOK, "%s", c.GetString("request_id"))
	})

	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	headerID := recorder.Header().Get("X-Request-ID")
	if headerID == "" {
		t.Fatal("expected non-empty X-Request-ID header")
	}
	if recorder.Body.String() != headerID {
		t.Fatalf("expected body request id %s, got %s", headerID, recorder.Body.String())
	}
}

func TestRequestIDMiddleware_ReuseHeader(t *testing.T) {
	engine := New()
	engine.Use(RequestID())
	engine.GET("/ping", func(c *Context) {
		c.String(http.StatusOK, "%s", c.GetString("request_id"))
	})

	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Request-ID", "custom-id-1")
	engine.ServeHTTP(recorder, req)

	if got := recorder.Header().Get("X-Request-ID"); got != "custom-id-1" {
		t.Fatalf("expected custom header id, got %s", got)
	}
}
