# GoFoundry

GoFoundry 是一个以学习驱动、工程化实现为目标的 Go 基础组件项目。

当前包含 7 个模块：
- `GoGee`：轻量 Web 框架
- `GoGorm`：轻量 ORM
- `GoCache`：分布式缓存
- `GoLock`：Redis 分布式锁
- `GoMQ`：RabbitMQ 简易封装
- `GoBench`：压测与数据采集程序（新增）
- `GoRPC`：预留模块（暂未实现）

## 1. 项目目标

把常见基础能力拆成可独立运行、可测试、可扩展的模块：
- Web 路由与中间件
- ORM 与事务
- 分布式缓存
- 分布式锁
- 消息队列
- 性能数据采集与报告

## 2. 模块能力

### 2.1 GoGee
- 路由分组与动态参数（`:id`、`*filepath`）
- 请求链控制（`Next`、`Abort`、`Fail`）
- JSON 绑定（`BindJSON`）
- 模板渲染（`SetFuncMap`、`LoadHTMLGlob`、`HTMLTemplate`）
- `NoRoute` / `NoMethod`
- `GET/POST/PUT/DELETE/PATCH/HEAD/OPTIONS/Any`

### 2.2 GoGorm
- 模型映射、CRUD、Where/OrderBy/Limit/Count
- 事务：`Begin/Commit/Rollback`
- 事务封装：`Engine.Transaction(...)`
- 上下文执行：`WithContext(ctx)`
- 字段选择：`Select(...)`
- Hook：`BeforeInsert/AfterInsert/BeforeQuery/AfterQuery`
- 自动迁移：`AutoMigrate()`
- 语义错误：`ErrRecordNotFound`

### 2.3 GoCache
- LRU + 一致性哈希 + HTTP 节点拉取
- Singleflight 防击穿
- TTL、批量读取/删除（`GetMany`、`RemoveMany`）
- 运行统计（命中、未命中、加载、淘汰、缓存大小）
- HTTPPool 可配置（副本数、basePath、client）

API 示例：
- `GET /api?key=Tom`
- `GET /api/batch?keys=Tom,Sam`
- `GET /api/stats`
- `GET /api/healthz`
- `DELETE /api/delete?key=Tom`

### 2.4 GoLock
- `TryLock` / `Lock`（等待超时）
- `Lease.Unlock` / `Lease.Refresh` / `Lease.KeepAlive`
- `Do(...)` 自动加锁执行
- Lua + owner token 原子校验（避免误删他人锁）

### 2.5 GoMQ
- `Dial` / `Publish` / `PublishJSON`
- `Consume` / `ConsumeWorkers` / `Close`
- exchange/queue 声明与 bind
- QoS（prefetch）
- ack/nack/requeue

### 2.6 GoBench（新增）
用于给项目产出可量化数据（吞吐、延迟分位数、成功率等），支持两种模式：
- `http`：压测任意 HTTP 接口（适合 GoGee/GoCache）
- `gorm`：本地 SQLite 下的 ORM 操作基准（Insert/Query/Update/Delete/Count）

输出：
- 控制台摘要（QPS、成功率、P95/P99）
- JSON 报告文件（默认 `reports/<name>-<mode>.json`）

### 2.7 GoRPC
- 预留模块，暂未实现。

## 3. 目录结构

```text
GoFoundry/
  GoGee/
  GoGorm/
  GoCache/
  GoLock/
  GoMQ/
  GoBench/
  GoRPC/
  LICENSE
  THIRD_PARTY_NOTICES.md
  go.work
```

## 4. 环境要求

- Go `1.24+`
- Redis（运行 GoLock 示例时需要）
- RabbitMQ（运行 GoMQ 示例时需要）

如果本地处于 GOPATH 模式：

```powershell
$env:GO111MODULE='on'
```

## 5. 快速运行

### 5.1 GoGee

```powershell
cd GoGee
go run .
```

### 5.2 GoGorm

```powershell
cd GoGorm
go run .
```

### 5.3 GoCache（三节点）

终端 1：

```powershell
cd GoCache
go run . -port=8001 -api=true
```

终端 2：

```powershell
cd GoCache
go run . -port=8002
```

终端 3：

```powershell
cd GoCache
go run . -port=8003
```

### 5.4 GoLock

```powershell
cd GoLock
go run . -redis=localhost:6379 -key=order-1001 -hold=5s
```

### 5.5 GoMQ

发布：

```powershell
cd GoMQ
go run . -mode=publish -queue=demo.queue -msg="hello" -exchange=demo.exchange -routing-key=demo.queue
```

消费：

```powershell
cd GoMQ
go run . -mode=consume -queue=demo.queue
```

## 6. 数据采集（GoBench）

### 6.1 HTTP 压测示例（压 GoGee / GoCache）

```powershell
cd GoBench
go run . -mode=http -name=gee-health -url=http://localhost:9999/healthz -method=GET -total=3000 -concurrency=100 -timeout=3s
```

### 6.2 GORM 基准示例

```powershell
cd GoBench
go run . -mode=gorm -name=gorm-crud -db=bench_gorm.db -insert=2000 -query=500 -update=500 -delete=500
```

### 6.3 指定报告路径

```powershell
go run . -mode=http -name=cache-batch -url=http://localhost:9999/api/batch?keys=Tom,Sam -report=reports/cache-batch.json
```

## 7. 测试

```powershell
$env:GO111MODULE='on'

cd GoGee
go test ./...

cd ..\GoGorm
go test ./...

cd ..\GoCache
go test ./...

cd ..\GoLock
go test ./...

cd ..\GoMQ
go test ./...

cd ..\GoBench
go test ./...
```

## 8. 开源来源与合规说明

`GoGee`、`GoGorm`、`GoCache` 的设计与部分实现思路参考：
- https://github.com/geektutu/7days-golang

许可证与声明已保留：
- `LICENSE`
- `THIRD_PARTY_NOTICES.md`

## 9. 后续计划

- 完成 `GoRPC` 模块
- GoGee 增强参数校验与路由命名
- GoGorm 增强 Scopes / 批量更新
- GoCache 增加指标导出（Prometheus）
- GoLock 增加 etcd 实现
- GoMQ 增加重试队列 / 延迟队列
- GoBench 增加 CSV 对比报告与基准历史归档