package log

import (
	"io"
	"log"
	"os"
	"sync"
)

var (
	errorLog = log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile)
	infoLog  = log.New(os.Stdout, "\033[34m[info]\033[0m ", log.LstdFlags|log.Lshortfile)
	loggers  = []*log.Logger{errorLog, infoLog}
	mux      sync.Mutex
)
var (
	Error  = errorLog.Println
	ErrorF = errorLog.Printf
	Info   = infoLog.Println
	InfoF  = infoLog.Printf
)

const (
	InfoLevel = iota
	ErrorLevel
	Disabled
)

const (
	InfoLovel  = InfoLevel
	ErrorLovel = ErrorLevel
)

func SetLevel(level int) {
	mux.Lock()
	defer mux.Unlock()
	for _, logger := range loggers {
		//log.Println(logger)
		logger.SetOutput(os.Stdout)
	}
	if level >= InfoLevel {
		infoLog.SetOutput(io.Discard)
	}
	if level >= ErrorLevel {
		errorLog.SetOutput(io.Discard)
	}
}
