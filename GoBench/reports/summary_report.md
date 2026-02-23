# 五框架测试总览

生成时间：2026-02-23 21:27:48 +08:00

## 本次重跑范围
- 全量单元测试：GoGee / GoGorm / GoCache / GoLock / GoMQ / GoBench
- 全量数据测试：GoGee / GoCache / GoGorm / GoLock / GoMQ

## 最终结论
- GoGee：通过（HTTP 压测 + 单元测试）
- GoCache：通过（HTTP 压测 + 单元测试）
- GoGorm：通过（GORM 基准 + 单元测试）
- GoLock：通过（锁 benchmark + 单元测试）
- GoMQ：通过（消息 benchmark + 单元测试）

## 查看顺序（建议）
1. `summary_report.md`（先看总览）
2. `details/` 下的 5 个框架报告（看中文解释）
3. `data/` 与 `logs/`（看原始数据）

## 目录结构
- `details/gogee_report.md`
- `details/gocache_report.md`
- `details/gogorm_report.md`
- `details/golock_report.md`
- `details/gomq_report.md`
- `data/*.json`（结构化测试数据）
- `logs/unit_test_all.log`（全量单元测试原始输出）
- `logs/golock_benchmark.log`、`logs/gomq_benchmark.log`（benchmark 原始输出）
