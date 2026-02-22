package Gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

func Recovery() HandlerFunc {
	return func(context *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%v", err)
				log.Print(trace(message))
				context.JSON(http.StatusInternalServerError, H{"message": "internal server error"})
			}
		}()
		context.Next()
	}
}
