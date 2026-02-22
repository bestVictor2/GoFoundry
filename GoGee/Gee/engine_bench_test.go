package Gee

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkEngineHealthz(b *testing.B) {
	engine := New()
	engine.GET("/healthz", func(c *Context) {
		c.JSON(http.StatusOK, H{"status": "ok"})
	})

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
		engine.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status %d", rec.Code)
		}
	}
}
