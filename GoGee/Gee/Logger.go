package Gee

import (
	"log"
	"time"
)

func Logger() HandlerFunc {
	return func(c *Context) {
		t := time.Now()
		c.Next()
		requestID := c.GetString("request_id")
		if requestID != "" {
			log.Printf("[%d] %s %s in %v", c.StatusCode, requestID, c.Rep.RequestURI, time.Since(t))
			return
		}
		log.Printf("[%d] %s in %v", c.StatusCode, c.Rep.RequestURI, time.Since(t))
	}
}
