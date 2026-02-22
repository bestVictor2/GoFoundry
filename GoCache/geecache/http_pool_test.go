package geecache

import (
	"net/http"
	"testing"
	"time"
)

func TestHTTPPoolOptions(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	pool := NewHTTPPoolWithOptions(
		"http://localhost:8001",
		WithHTTPPoolBasePath("cache"),
		WithHTTPPoolReplicas(77),
		WithHTTPPoolClient(client),
	)
	if pool.basePath != "/cache/" {
		t.Fatalf("expected normalized base path /cache/, got %s", pool.basePath)
	}
	if pool.replicas != 77 {
		t.Fatalf("expected replicas=77, got %d", pool.replicas)
	}
	if pool.client != client {
		t.Fatal("expected custom client configured")
	}
}
