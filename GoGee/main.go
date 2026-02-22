package main

import (
	"GoGee/Gee"
	"log"
	"net/http"
	"time"
)

func traceByGroup(group string) Gee.HandlerFunc {
	return func(c *Gee.Context) {
		t := time.Now()
		c.Next()
		log.Printf("[%d] %s in %v for group %s", c.StatusCode, c.Rep.RequestURI, time.Since(t), group)
	}
}

func main() {
	r := Gee.Default()
	r.Use(Gee.RequestID())
	r.NoRoute(func(c *Gee.Context) {
		c.JSON(http.StatusNotFound, Gee.H{
			"error": "route not found",
			"path":  c.Path,
		})
	})
	r.NoMethod(func(c *Gee.Context) {
		c.JSON(http.StatusMethodNotAllowed, Gee.H{
			"error":  "method not allowed",
			"method": c.Method,
			"path":   c.Path,
		})
	})

	r.GET("/", func(c *Gee.Context) {
		c.HTML(http.StatusOK, "<h1>GoFoundry / GoGee</h1>")
	})
	r.Any("/ping", func(c *Gee.Context) {
		c.JSON(http.StatusOK, Gee.H{
			"message": "pong",
			"method":  c.Method,
		})
	})
	r.GET("/healthz", func(c *Gee.Context) {
		c.JSON(http.StatusOK, Gee.H{
			"status":     "ok",
			"service":    "go-gee",
			"request_id": c.GetString("request_id"),
			"time":       time.Now().Format(time.RFC3339),
		})
	})

	api := r.Group("/api")
	v1 := api.Group("/v1")
	v1.GET("/hello/:name", func(c *Gee.Context) {
		c.JSON(http.StatusOK, Gee.H{
			"message":    "hello " + c.Param("name"),
			"path":       c.Path,
			"request_id": c.GetString("request_id"),
		})
	})
	v1.POST("/echo", func(c *Gee.Context) {
		type payload struct {
			Message string `json:"message"`
		}
		var body payload
		if err := c.BindJSON(&body); err != nil {
			c.Fail(http.StatusBadRequest, "invalid json body")
			return
		}
		c.JSON(http.StatusOK, Gee.H{
			"echo":       body.Message,
			"request_id": c.GetString("request_id"),
		})
	})

	v2 := r.Group("/v2")
	v2.Use(traceByGroup("v2"))
	v2.GET("/hello/:name", func(c *Gee.Context) {
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
	})
	v2.PUT("/profile/:name", func(c *Gee.Context) {
		c.JSON(http.StatusOK, Gee.H{
			"updated": c.Param("name"),
			"query":   c.Query("source"),
		})
	})

	r.GET("/__routes", func(c *Gee.Context) {
		c.JSON(http.StatusOK, Gee.H{"routes": r.Routes()})
	})

	r.Run(":9999")
}
