package main

import (
	"GoGorm/gorm"
	gLog "GoGorm/log"
	"GoGorm/session"
	"context"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type GormBenchConfig struct {
	DBPath    string
	InsertOps int
	QueryOps  int
	UpdateOps int
	DeleteOps int
}

type benchUser struct {
	Name string `foundry:"PRIMARY KEY"`
	Age  int
	City string
}

func runGormBenchmark(name, reportPath string, cfg GormBenchConfig) error {
	if cfg.DBPath == "" {
		cfg.DBPath = "bench_gorm.db"
	}
	if cfg.InsertOps <= 0 {
		cfg.InsertOps = 1000
	}
	if cfg.QueryOps < 0 {
		cfg.QueryOps = 0
	}
	if cfg.UpdateOps < 0 {
		cfg.UpdateOps = 0
	}
	if cfg.DeleteOps < 0 {
		cfg.DeleteOps = 0
	}
	if cfg.DeleteOps > cfg.InsertOps {
		cfg.DeleteOps = cfg.InsertOps
	}
	gLog.SetLevel(gLog.InfoLevel)

	_ = os.Remove(cfg.DBPath)

	engine, err := gorm.NewEngine("sqlite3", cfg.DBPath)
	if err != nil {
		return err
	}
	defer engine.Close()

	s := engine.NewSession().WithContext(context.Background()).Model(&benchUser{})
	if err = s.DropTable(); err != nil {
		return err
	}
	if err = s.AutoMigrate(); err != nil {
		return err
	}

	stageDuration := map[string]int64{}
	errorStats := map[string]int{}
	latencies := make([]time.Duration, 0, cfg.InsertOps+cfg.QueryOps+cfg.UpdateOps+cfg.DeleteOps+1) // 每种操作延迟
	success := 0
	failed := 0

	record := func(op string, lat time.Duration, err error) {
		stageDuration[op] += lat.Milliseconds()
		latencies = append(latencies, lat)
		if err != nil {
			failed++
			errorStats[op]++
		} else {
			success++
		}
	}

	startAll := time.Now()

	for i := 0; i < cfg.InsertOps; i++ { // insert 操作
		u := &benchUser{Name: fmt.Sprintf("user_%06d", i), Age: i % 80, City: "shanghai"}
		start := time.Now()
		_, err = s.Insert(u)
		record("insert", time.Since(start), err)
	}

	for i := 0; i < cfg.QueryOps; i++ { // query 操作
		target := fmt.Sprintf("user_%06d", i%cfg.InsertOps)
		var u benchUser
		start := time.Now()
		err = s.Where("Name = ?", target).First(&u)
		record("query", time.Since(start), err)
	}

	for i := 0; i < cfg.UpdateOps; i++ { // update 操作
		target := fmt.Sprintf("user_%06d", i%cfg.InsertOps)
		start := time.Now()
		_, err = s.Where("Name = ?", target).Update("Age", 100+i%20)
		record("update", time.Since(start), err)
	}

	for i := 0; i < cfg.DeleteOps; i++ { // delete 操作
		target := fmt.Sprintf("user_%06d", i)
		start := time.Now()
		_, err = s.Where("Name = ?", target).Delete()
		record("delete", time.Since(start), err)
	}

	// count 阶段
	startCount := time.Now()
	remaining, countErr := s.Count()
	record("count", time.Since(startCount), countErr)

	endAll := time.Now()
	total := success + failed
	duration := endAll.Sub(startAll)
	successRate := 0.0
	if total > 0 {
		successRate = float64(success) / float64(total)
	}
	qps := 0.0
	if duration > 0 {
		qps = float64(total) / duration.Seconds()
	}
	avg, min, p50, p90, p95, p99, max := buildLatencySummary(latencies)

	report := Report{
		Name:       name,
		Mode:       "gorm",
		StartAt:    startAll.Format(time.RFC3339),
		EndAt:      endAll.Format(time.RFC3339),
		DurationMS: duration.Milliseconds(),
		Summary: Summary{
			Total:        total,
			Success:      success,
			Failed:       failed,
			SuccessRate:  successRate,
			QPS:          qps,
			AvgLatencyMS: avg,
			MinLatencyMS: min,
			P50LatencyMS: p50,
			P90LatencyMS: p90,
			P95LatencyMS: p95,
			P99LatencyMS: p99,
			MaxLatencyMS: max,
			ErrorStat:    errorStats,
		},
		Extra: map[string]any{
			"db_path":                 cfg.DBPath,
			"insert_ops":              cfg.InsertOps,
			"query_ops":               cfg.QueryOps,
			"update_ops":              cfg.UpdateOps,
			"delete_ops":              cfg.DeleteOps,
			"remaining_rows":          remaining,
			"stage_ms":                stageDuration,
			"record_not_found_symbol": session.ErrRecordNotFound.Error(),
		},
	}
	if err = writeReport(reportPath, report); err != nil {
		return err
	}
	printSummary(report, reportPath)
	return nil
}
