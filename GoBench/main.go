package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"
)

func main() {
	var (
		mode   string
		name   string
		report string

		httpCfg HTTPBenchConfig
		gormCfg GormBenchConfig
	)

	flag.StringVar(&mode, "mode", "http", "benchmark mode: http | gorm")
	flag.StringVar(&name, "name", "benchmark", "benchmark scenario name")
	flag.StringVar(&report, "report", "", "report output path, default reports/<name>-<mode>.json")

	flag.StringVar(&httpCfg.URL, "url", "http://localhost:9999/healthz", "http mode: request url")
	flag.StringVar(&httpCfg.Method, "method", "GET", "http mode: request method")
	flag.IntVar(&httpCfg.Total, "total", 2000, "http mode: total requests")
	flag.IntVar(&httpCfg.Concurrency, "concurrency", 100, "http mode: concurrent workers")
	flag.DurationVar(&httpCfg.Timeout, "timeout", 3*time.Second, "http mode: request timeout")
	flag.IntVar(&httpCfg.Warmup, "warmup", 50, "http mode: warmup requests")
	flag.StringVar(&httpCfg.Body, "body", "", "http mode: request body")
	flag.StringVar(&httpCfg.HeaderText, "headers", "", "http mode: headers, e.g. k1=v1,k2=v2")

	flag.StringVar(&gormCfg.DBPath, "db", "bench_gorm.db", "gorm mode: sqlite db file path")
	flag.IntVar(&gormCfg.InsertOps, "insert", 2000, "gorm mode: insert operations")
	flag.IntVar(&gormCfg.QueryOps, "query", 500, "gorm mode: query operations")
	flag.IntVar(&gormCfg.UpdateOps, "update", 500, "gorm mode: update operations")
	flag.IntVar(&gormCfg.DeleteOps, "delete", 500, "gorm mode: delete operations")
	flag.Parse()

	if report == "" {
		report = fmt.Sprintf("reports/%s-%s.json", sanitizeFileName(name), mode)
	}

	var err error
	switch strings.ToLower(mode) {
	case "http":
		err = runHTTPBenchmark(name, report, httpCfg)
	case "gorm":
		err = runGormBenchmark(name, report, gormCfg)
	default:
		err = errors.New("unsupported mode, use http or gorm")
	}
	if err != nil {
		log.Fatal(err)
	}
}
