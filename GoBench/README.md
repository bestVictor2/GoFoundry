# GoBench 使用说明

## 目标
`GoBench` 用于统一输出 GoFoundry 五个框架的测试结果：
- GoGee / GoCache：HTTP 压测数据
- GoGorm：ORM CRUD 基准数据
- GoLock / GoMQ：Benchmark 数据
- 全模块单元测试汇总日志

## 本次结果位置
所有结果都在 `GoBench/reports/`：
- 总览：`GoBench/reports/summary_report.md`
- 目录说明：`GoBench/reports/reports_readme.md`
- 各框架中文报告：`GoBench/reports/details/*_report.md`
- 结构化数据：`GoBench/reports/data/*.json`
- 原始日志：`GoBench/reports/logs/*.log`

## 手动重跑（推荐顺序）
1. 进入 GoBench：
```powershell
cd GoBench
```
2. 先跑单元测试（五框架 + GoBench）：
```powershell
cd ..\GoGee;   go test -count=1 ./...
cd ..\GoGorm;  go test -count=1 ./...
cd ..\GoCache; go test -count=1 ./...
cd ..\GoLock;  go test -count=1 ./...
cd ..\GoMQ;    go test -count=1 ./...
cd ..\GoBench; go test -count=1 ./...
```
3. 跑 GoGorm 数据测试：
```powershell
cd ..\GoBench
go run . -mode=gorm -name=gogorm-benchmark -db=reports/data/gogorm_bench.db -insert=2000 -query=500 -update=500 -delete=500
```
4. 启动 GoGee 后跑 GoGee 数据测试：
```powershell
cd ..\GoGee
go run .
# 新终端
cd ..\GoBench
go run . -mode=http -name=gogee-benchmark -url=http://localhost:9999/healthz -method=GET -total=3000 -concurrency=100 -timeout=3s -warmup=50
```
5. 启动 GoCache 三节点后跑 GoCache 数据测试：
```powershell
cd ..\GoCache
go run . -port=8001 -api=true
# 新终端
go run . -port=8002
# 新终端
go run . -port=8003
# 新终端
cd ..\GoBench
go run . -mode=http -name=gocache-get-benchmark -url=http://localhost:9999/api?key=Tom -method=GET -total=3000 -concurrency=30 -timeout=10s -warmup=100
go run . -mode=http -name=gocache-batch-benchmark -url=http://localhost:9999/api/batch?keys=Tom,Sam -method=GET -total=3000 -concurrency=30 -timeout=10s -warmup=100
```
6. 跑 GoLock / GoMQ benchmark：
```powershell
cd ..\GoLock
go test -run ^$ -bench . -benchmem ./lock

cd ..\GoMQ
go test -run ^$ -bench . -benchmem ./mq
```

## 说明
- `GoBench` 当前只内置 `http` 与 `gorm` 两种模式，因此 GoLock/GoMQ 的数据测试通过 benchmark 方式补齐。
- 若需要真实中间件链路压测（如真实 RabbitMQ 网络 I/O），建议在现有 benchmark 基础上再加集成场景。
- 1.GoGee：大概率比 Gin/Echo 在同机同压测下低一个明显档位，尤其 P95/P99 偏高。
- 2.GoCache：吞吐和延迟都还行，通常是“小到中等差距”。
- 3.GoGorm：在 SQLite 本地基准里不算差，属于“同量级可用”。
- 4.GoLock：因为用了 miniredis 同进程 benchmark，结果不等价真实 Redis 网络场景。
- 5.GoMQ：因为不是打真实 RabbitMQ broker，结果会比真实场景乐观很多，不能拿来和生产框架对标。
- 仍在改进 ing