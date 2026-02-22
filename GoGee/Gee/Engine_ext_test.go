package Gee

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEngineAnyRoute(t *testing.T) {
	engine := New()
	engine.Any("/any", func(c *Context) {
		c.String(http.StatusOK, "ok")
	})

	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/any", nil)
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK || recorder.Body.String() != "ok" {
		t.Fatalf("expected any route works, got status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestEngineNoRouteAndNoMethod(t *testing.T) {
	engine := New()
	engine.GET("/users/:name", func(c *Context) {
		c.String(http.StatusOK, "%s", c.Param("name"))
	})
	engine.NoRoute(func(c *Context) {
		c.String(http.StatusNotFound, "custom not found")
	})
	engine.NoMethod(func(c *Context) {
		c.String(http.StatusMethodNotAllowed, "custom method not allowed")
	})

	rec404 := httptest.NewRecorder()
	req404, _ := http.NewRequest(http.MethodGet, "/missing", nil)
	engine.ServeHTTP(rec404, req404)
	if rec404.Code != http.StatusNotFound || rec404.Body.String() != "custom not found" {
		t.Fatalf("unexpected no route response status=%d body=%s", rec404.Code, rec404.Body.String())
	}

	rec405 := httptest.NewRecorder()
	req405, _ := http.NewRequest(http.MethodPost, "/users/tom", nil)
	engine.ServeHTTP(rec405, req405)
	if rec405.Code != http.StatusMethodNotAllowed || rec405.Body.String() != "custom method not allowed" {
		t.Fatalf("unexpected no method response status=%d body=%s", rec405.Code, rec405.Body.String())
	}
	if allow := rec405.Header().Get("Allow"); !strings.Contains(allow, http.MethodGet) {
		t.Fatalf("expected Allow header contains GET, got %s", allow)
	}
}

func TestContextFailAbort(t *testing.T) {
	engine := New()
	called := false
	engine.Use(func(c *Context) {
		c.Fail(http.StatusUnauthorized, "blocked")
	})
	engine.GET("/secure", func(c *Context) {
		called = true
		c.String(http.StatusOK, "ok")
	})

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/secure", nil)
	engine.ServeHTTP(rec, req)

	if called {
		t.Fatal("handler should not be executed after Fail")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "blocked") {
		t.Fatalf("unexpected fail body: %s", rec.Body.String())
	}
}

func TestContextHTMLTemplate(t *testing.T) {
	engine := New()
	dir := t.TempDir()
	tplFile := filepath.Join(dir, "hello.tmpl")
	content := "hello {{upper .Name}}"
	if err := os.WriteFile(tplFile, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp template failed: %v", err)
	}

	engine.SetFuncMap(template.FuncMap{
		"upper": strings.ToUpper,
	})
	engine.LoadHTMLGlob(filepath.Join(dir, "*.tmpl"))
	engine.GET("/hello", func(c *Context) {
		c.HTMLTemplate(http.StatusOK, "hello.tmpl", H{"Name": "tom"})
	})

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/hello", nil)
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != "hello TOM" {
		t.Fatalf("unexpected template body: %s", rec.Body.String())
	}
}
