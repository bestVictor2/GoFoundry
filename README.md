# GoFoundry

GoFoundry 是一个以学习驱动、偏工程化实现的 Go 基础组件项目。

当前包含 6 个模块：
- `GoGee`：轻量 Web 框架
- `GoGorm`：轻量 ORM
- `GoCache`：分布式缓存
- `GoLock`：Redis 分布式锁
- `GoMQ`：RabbitMQ 简易封装
- `GoRPC`：预留模块（暂未实现）

## 1. 项目目标

这个仓库不是单一业务项目，而是把常见基础能力做成独立可运行模块：
- HTTP 路由与中间件
- ORM 与事务
- 分布式缓存
- 分布式锁
- 消息队列

目标是形成一套可学习、可扩展、可继续演进的基础设施代码库。

## 2. 模块说明

### 2.1 GoGee（轻量 Web 框架）

核心能力：
- 路由分组与中间件
- 动态路由参数（`:id`、`*filepath`）
- 请求链控制（`Next`、`Abort`、`Fail`）
- JSON 绑定（`BindJSON`）
- 自定义 404/405（`NoRoute`、`NoMethod`）
- 模板渲染（`SetFuncMap`、`LoadHTMLGlob`、`HTMLTemplate`）
- RequestID、日志、异常恢复中间件

支持方法：
- `GET/POST/PUT/DELETE/PATCH/HEAD/OPTIONS`
- `Any`

### 2.2 GoGorm（轻量 ORM）

核心能力：
- 模型映射与建表
- CRUD、Where、OrderBy、Limit、Count
- 事务（`Begin/Commit/Rollback`）
- 事务封装（`Engine.Transaction`）
- 上下文执行（`WithContext`）
- 字段选择（`Select("Name", "Age")`）
- Hook：`BeforeInsert/AfterInsert/BeforeQuery/AfterQuery`
- 自动迁移（`AutoMigrate`）
- `ErrRecordNotFound` 语义

### 2.3 GoCache（分布式缓存）

核心能力：
- 本地 LRU
- 一致性哈希 + 多节点 HTTP 拉取
- Singleflight 防击穿
- TTL 支持
- 批量读取与批量删除（`GetMany`、`RemoveMany`）
- 统计信息（命中、未命中、加载次数、淘汰、缓存大小）
- HTTPPool 可配置（副本数、basePath、client）

示例 API：
- `GET /api?key=Tom`
- `GET /api/batch?keys=Tom,Sam`
- `GET /api/stats`
- `GET /api/healthz`
- `DELETE /api/delete?key=Tom`

### 2.4 GoLock（Redis 分布式锁）

核心能力：
- `TryLock`
- `Lock`（带等待超时）
- `Lease.Unlock`
- `Lease.Refresh`
- `Lease.KeepAlive`
- `Do`（自动加锁执行并释放）

实现特性：
- 使用 owner token 标识锁持有者
- 解锁与续期走 Lua 原子脚本，避免误删他人锁

### 2.5 GoMQ（RabbitMQ 简易封装）

核心能力：
- `Dial`
- `Publish`
- `PublishJSON`
- `Consume`
- `ConsumeWorkers`
- `Close`

实现特性：
- exchange/queue 声明与 bind
- QoS（prefetch）
- ack/nack/requeue 处理
- 支持并发 worker 消费

### 2.6 GoRPC

当前状态：
- 预留模块，暂未实现。

## 3. 目录结构

```text
GoFoundry/
  GoGee/
  GoGorm/
  GoCache/
  GoLock/
  GoMQ/
  GoRPC/
  LICENSE
  THIRD_PARTY_NOTICES.md
  go.work
```

## 4. 环境要求

- Go `1.24+`
- Redis（运行 GoLock 示例时需要）
- RabbitMQ（运行 GoMQ 示例时需要）

如果本地处于 GOPATH 模式，先启用 module：

```powershell
$env:GO111MODULE='on'
```

## 5. 快速运行

### 5.1 GoGee

```powershell
cd GoGee
go run .
```

默认监听：`http://localhost:9999`

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

## 6. 测试

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
```

## 7. 开源来源与合规说明

本项目中 `GoGee`、`GoGorm`、`GoCache` `GoPRC` 的设计与部分实现思路参考了：
- https://github.com/geektutu/7days-golang

相关许可证信息与上游声明已在仓库中保留：
- `LICENSE`
- `THIRD_PARTY_NOTICES.md`

## 8. 后续计划

- 完成 `GoRPC` 模块
- 为 `GoGee` 增加更完整的参数校验与路由命名能力
- 为 `GoGorm` 增加更丰富的查询构建能力（Scopes、批量更新等）
- 为 `GoCache` 增加指标导出（如 Prometheus）
- 为 `GoLock` 增加 etcd 实现
- 为 `GoMQ` 增加重试队列 / 延迟队列支持