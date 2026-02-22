package Gee

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func newTestRouter() *Router {
	r := newRouter()
	r.addRouter("GET", "/", nil)
	r.addRouter("GET", "/hello/:name", nil)
	r.addRouter("GET", "/hello/b/c", nil)
	r.addRouter("GET", "/hi/:name", nil)
	r.addRouter("GET", "/assets/*filepath", nil)
	return r
}

func TestParsePattern(t *testing.T) {
	ok := reflect.DeepEqual(parsePattern("/p/:name"), []string{"p", ":name"})
	ok = ok && reflect.DeepEqual(parsePattern("/p/*"), []string{"p", "*"})
	ok = ok && reflect.DeepEqual(parsePattern("/p/*name/*"), []string{"p", "*name"})
	if !ok {
		t.Fatal("test parsePattern failed")
	}
}

func TestGetRoute(t *testing.T) {
	r := newTestRouter()
	n, ps := r.getRouter("GET", "/hello/geektutu")

	if n == nil {
		t.Fatal("nil shouldn't be returned")
	}

	if n.pattern != "/hello/:name" {
		t.Fatal("should match /hello/:name")
	}

	if ps["name"] != "geektutu" {
		t.Fatal("name should be equal to 'geektutu'")
	}

	fmt.Printf("matched path: %s, params['name']: %s\n", n.pattern, ps["name"])

}

func TestNestedGroupPrefix(t *testing.T) {
	engine := New()
	v1 := engine.Group("/api").Group("/v1")
	v1.GET("/ping", func(c *Context) {
		c.String(http.StatusOK, "pong")
	})

	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if recorder.Body.String() != "pong" {
		t.Fatalf("expected body pong, got %s", recorder.Body.String())
	}
}

func TestListRoutes(t *testing.T) {
	engine := New()
	engine.GET("/a", func(c *Context) {})
	engine.POST("/b", func(c *Context) {})
	routes := engine.Routes()
	if len(routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(routes))
	}
}
