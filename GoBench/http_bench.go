package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type HTTPBenchConfig struct {
	URL         string
	Method      string
	Total       int
	Concurrency int
	Timeout     time.Duration
	Warmup      int // 预热请求数
	Body        string
	HeaderText  string
}

type httpResult struct { // 单次请求结构体
	latency time.Duration
	status  int
	ok      bool
	err     string
}

func runHTTPBenchmark(name, reportPath string, cfg HTTPBenchConfig) error {
	if cfg.Total <= 0 {
		cfg.Total = 1
	}
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 1
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 3 * time.Second
	}
	if cfg.Method == "" {
		cfg.Method = http.MethodGet
	}
	cfg.Method = strings.ToUpper(cfg.Method)

	headers := parseHeaders(cfg.HeaderText)
	client := &http.Client{Timeout: cfg.Timeout}

	for i := 0; i < cfg.Warmup; i++ { // 预热请求
		_ = doOneHTTP(client, cfg, headers)
	}

	results := make(chan httpResult, cfg.Total)
	jobs := make(chan struct{}, cfg.Total) // 并发 jobs 数量

	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for range jobs {
			results <- doOneHTTP(client, cfg, headers)
		}
	}
	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go worker()
	}

	start := time.Now()
	for i := 0; i < cfg.Total; i++ { // 统计纯执行时间
		jobs <- struct{}{}
	}
	close(jobs)
	wg.Wait()
	close(results)
	end := time.Now()

	var (
		success     int
		failed      int
		latencies   []time.Duration // 延时统计
		statusStats = make(map[int]int)
		errorStats  = make(map[string]int)
	)
	for result := range results {
		if result.status != 0 {
			statusStats[result.status]++
		}
		if result.err != "" {
			errorStats[result.err]++
		}
		if result.ok {
			success++
			latencies = append(latencies, result.latency)
		} else {
			failed++
		}
	}

	duration := end.Sub(start)
	avg, min, p50, p90, p95, p99, max := buildLatencySummary(latencies)
	total := success + failed
	successRate := 0.0
	if total > 0 {
		successRate = float64(success) / float64(total)
	}
	qps := 0.0 // qps 平均每秒请求数量
	if duration > 0 {
		qps = float64(total) / duration.Seconds()
	}

	report := Report{
		Name:       name,
		Mode:       "http",
		StartAt:    start.Format(time.RFC3339),
		EndAt:      end.Format(time.RFC3339),
		DurationMS: duration.Milliseconds(),
		Summary: Summary{
			Total:          total,
			Success:        success,
			Failed:         failed,
			SuccessRate:    successRate,
			QPS:            qps,
			AvgLatencyMS:   avg,
			MinLatencyMS:   min,
			P50LatencyMS:   p50,
			P90LatencyMS:   p90,
			P95LatencyMS:   p95,
			P99LatencyMS:   p99,
			MaxLatencyMS:   max,
			StatusCodeStat: statusStats,
			ErrorStat:      errorStats,
		},
		Extra: map[string]any{
			"url":         cfg.URL,
			"method":      cfg.Method,
			"total":       cfg.Total,
			"concurrency": cfg.Concurrency,
			"timeout_ms":  cfg.Timeout.Milliseconds(),
			"warmup":      cfg.Warmup,
		},
	}

	if err := writeReport(reportPath, report); err != nil {
		return err
	}
	printSummary(report, reportPath)
	return nil
}

func doOneHTTP(client *http.Client, cfg HTTPBenchConfig, headers map[string]string) httpResult {
	body := strings.NewReader(cfg.Body) // 包装为一个 io.reader
	req, err := http.NewRequestWithContext(context.Background(), cfg.Method, cfg.URL, body)
	if err != nil {
		return httpResult{ok: false, err: "build_request_error"}
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start)
	if err != nil {
		return httpResult{latency: latency, ok: false, err: "request_error"}
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body) // 不读完 body qps 会降低 原因是在需要建立 TCP 请求 如果不读取 body 的话会重新建立 tpc 连接

	ok := resp.StatusCode >= 200 && resp.StatusCode < 400
	errCode := ""
	if !ok {
		errCode = fmt.Sprintf("status_%d", resp.StatusCode)
	}
	return httpResult{
		latency: latency,
		status:  resp.StatusCode,
		ok:      ok,
		err:     errCode,
	}
}

func parseHeaders(text string) map[string]string {
	headers := make(map[string]string)
	if strings.TrimSpace(text) == "" {
		return headers
	}
	pairs := strings.Split(text, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key != "" {
			headers[key] = value
		}
	}
	return headers
}
