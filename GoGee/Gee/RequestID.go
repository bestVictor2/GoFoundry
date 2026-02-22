package Gee

import (
	"strconv"
	"sync/atomic"
	"time"
)

var requestIDSeed uint64

func nextRequestID() string {
	seq := atomic.AddUint64(&requestIDSeed, 1)
	ts := strconv.FormatInt(time.Now().UnixNano(), 36)
	return ts + "-" + strconv.FormatUint(seq, 36)
}

func RequestID() HandlerFunc {
	return func(c *Context) {
		requestID := c.Rep.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = nextRequestID()
		}
		c.Set("request_id", requestID)
		c.SetHeader("X-Request-ID", requestID)
		c.Next()
	}
}
