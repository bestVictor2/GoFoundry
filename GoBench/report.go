package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Report struct {
	Name       string         `json:"name"`
	Mode       string         `json:"mode"`
	StartAt    string         `json:"start_at"`
	EndAt      string         `json:"end_at"`
	DurationMS int64          `json:"duration_ms"`
	Summary    Summary        `json:"summary"`
	Extra      map[string]any `json:"extra,omitempty"`
}

type Summary struct {
	Total          int            `json:"total"`
	Success        int            `json:"success"`
	Failed         int            `json:"failed"`
	SuccessRate    float64        `json:"success_rate"`
	QPS            float64        `json:"qps"`
	AvgLatencyMS   float64        `json:"avg_latency_ms,omitempty"`
	MinLatencyMS   float64        `json:"min_latency_ms,omitempty"`
	P50LatencyMS   float64        `json:"p50_latency_ms,omitempty"`
	P90LatencyMS   float64        `json:"p90_latency_ms,omitempty"`
	P95LatencyMS   float64        `json:"p95_latency_ms,omitempty"`
	P99LatencyMS   float64        `json:"p99_latency_ms,omitempty"`
	MaxLatencyMS   float64        `json:"max_latency_ms,omitempty"`
	StatusCodeStat map[int]int    `json:"status_code_stat,omitempty"`
	ErrorStat      map[string]int `json:"error_stat,omitempty"`
}

func buildLatencySummary(latencies []time.Duration) (avg, min, p50, p90, p95, p99, max float64) {
	if len(latencies) == 0 {
		return
	}
	values := make([]float64, len(latencies))
	total := 0.0
	for i, lat := range latencies {
		ms := float64(lat.Microseconds()) / 1000.0 // ms
		values[i] = ms
		total += ms
	}
	sort.Float64s(values)
	avg = total / float64(len(values))
	min = values[0]
	max = values[len(values)-1]
	p50 = percentile(values, 50)
	p90 = percentile(values, 90)
	p95 = percentile(values, 95)
	p99 = percentile(values, 99)
	return
}

func percentile(sorted []float64, p int) float64 { // 百分位算法
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[len(sorted)-1]
	}
	idx := int(float64(len(sorted)-1) * (float64(p) / 100.0))
	return sorted[idx]
}

func writeReport(path string, report Report) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	return nil
}

func printSummary(report Report, reportPath string) {
	s := report.Summary
	fmt.Printf("=== %s (%s) ===\n", report.Name, strings.ToUpper(report.Mode))
	fmt.Printf("duration: %dms | total: %d | success: %d | failed: %d | success_rate: %.2f%%\n",
		report.DurationMS, s.Total, s.Success, s.Failed, s.SuccessRate*100)
	fmt.Printf("qps: %.2f | avg/p95/p99 latency(ms): %.2f / %.2f / %.2f\n",
		s.QPS, s.AvgLatencyMS, s.P95LatencyMS, s.P99LatencyMS)
	fmt.Printf("report: %s\n", reportPath)
}

func sanitizeFileName(name string) string { // 不同平台的安全设计
	name = strings.TrimSpace(name)
	if name == "" {
		return "benchmark"
	}
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	return strings.ToLower(name)
}
