# GoFoundry 项目面试问答（含答案）

说明：问题围绕你简历中的 GoFoundry 项目设计与实现，答案为可用于口述的简洁版本。

## 1. 项目总览与架构

1. Q：GoFoundry 的整体架构是什么？
A：以 Go 模块拆分成 Web（GoGee）、ORM（GoGorm）、缓存（GoCache）、分布式锁（GoLock）、消息队列（GoMQ）与 RPC（GoRPC），对外各自独立可运行、可测试，对内只通过少量通用基础包与接口依赖，减少强耦合。

2. Q：为什么要做基础组件库，而不是直接用成熟框架？
A：目标是掌握底层原理与工程化能力，针对特定业务裁剪功能，减少不必要的依赖与复杂度，同时验证性能与可维护性。

3. Q：你如何保证模块化同时不牺牲易用性？
A：对外提供稳定、简洁的入口 API；对内通过接口分层与默认实现封装复杂度；同时给出样例与测试基线。

4. Q：模块之间如何避免循环依赖？
A：抽象公共能力（日志、配置、上下文、错误模型）到基础包；高层模块依赖低层接口，不反向依赖具体实现。

5. Q：可独立运行、可测试体现在哪？
A：每个模块都有独立示例与测试入口，依赖注入连接/存储，保证单测可用 mock 替换外部系统。

## 2. Web 框架（GoGee）

6. Q：路由分组的实现思路？
A：通过 Group 持有前缀与中间件链，注册时拼接前缀与路由树节点，命中时合并中间件。

7. Q：路由匹配的数据结构？
A：前缀树（Trie），节点保存静态段与参数段，匹配时按段遍历并记录参数。

8. Q：静态路由与参数路由冲突如何处理？
A：优先匹配静态段，其次参数段；注册时可做冲突检测，避免同一路径多种匹配导致不确定性。

9. Q：中间件链如何实现？
A：通过切片保存处理函数，Context 维护索引，Next() 逐个执行形成责任链。

10. Q：NoRoute/NoMethod 的触发时机？
A：路由未命中或方法不匹配时触发；先匹配路径，再判定方法是否存在。

11. Q：Context 为什么要自定义？
A：封装请求/响应、路径参数、中间件状态，提升可读性和扩展性，同时减少重复代码。

12. Q：模板渲染如何实现？
A：加载模板文件为模板集合，渲染时结合数据与模板名输出到响应；支持自定义函数。

13. Q：如何处理 panic？
A：在最外层中间件做 recover，记录错误并返回统一错误响应，防止进程崩溃。

14. Q：如何支持请求取消与超时？
A：透传 request.Context()，在业务和下游调用中检查 Done 通道。

15. Q：路由性能瓶颈在哪里？
A：主要是路由树遍历与参数解析；优化点在节点结构、减少内存分配与字符串切分。

## 3. ORM（GoGorm）

16. Q：事务如何实现？
A：封装 Begin/Commit/Rollback，并在上下文中携带事务连接，确保同一事务内复用连接。

17. Q：支持嵌套事务吗？
A：简化版通常不支持，或通过计数器实现逻辑嵌套，只有最外层提交/回滚。

18. Q：Hook 的调用顺序？
A：一般是 BeforeCreate/AfterCreate、BeforeUpdate/AfterUpdate 等，按操作生命周期固定触发。

19. Q：AutoMigrate 如何保证幂等？
A：先读取现有表结构，对比字段和索引差异，仅补充缺失字段或索引。

20. Q：字段选择如何实现？
A：通过反射字段标签构建列名，Select/Omits 过滤字段集合，生成 SQL 列表。

21. Q：如何避免 SQL 注入？
A：SQL 结构由模板构建，参数值使用预编译占位符传入，不直接拼接值。

22. Q：反射性能如何优化？
A：缓存结构体元信息（字段名、标签、类型），避免重复反射开销。

23. Q：上下文执行模型是什么？
A：DB 操作接收 context.Context，执行时绑定到 sql.DB 或事务，支持超时/取消。

24. Q：错误模型如何设计？
A：区分可重试错误（临时网络、超时）与不可重试错误（语法、约束），对外暴露统一错误类型。

## 4. GoCache（一致性哈希 + singleflight + TTL）

25. Q：一致性哈希的核心优势？
A：节点增删时只影响少量 key，减少大规模迁移。

26. Q：虚拟节点如何选择数量？
A：与节点数和数据分布有关，通常几十到上百个虚拟节点以改善均匀性。

27. Q：singleflight 的作用？
A：多个并发请求命中同一缓存缺失时合并为一次回源，减少压力与抖动。

28. Q：TTL 过期策略？
A：读时惰性检查过期并删除，可选定时清理协程减少堆积。

29. Q：如何解决缓存击穿？
A：使用互斥锁或 singleflight 合并回源，热点 key 设置不过期或延长 TTL。

30. Q：如何处理缓存雪崩？
A：TTL 加随机抖动、分层缓存与限流降级。

31. Q：如何处理缓存穿透？
A：对不存在的数据缓存空值或使用布隆过滤器。

## 5. GoLock（Redis Lua 原子解锁/续期）

32. Q：为什么解锁要用 Lua？
A：保证“检查 token 是否一致”和“删除锁”是原子操作，避免误删。

33. Q：续期机制怎么做？
A：定时任务在锁持有期内刷新 TTL，并确认 token 一致后续期。

34. Q：续期失败怎么办？
A：认为锁已失效，终止业务或进入补偿流程，避免写入脏数据。

35. Q：如何防止误删别人锁？
A：每个锁带唯一 token，解锁 Lua 脚本验证 token 一致性。

36. Q：分布式锁的可靠性边界？
A：依赖 Redis 单点或主从一致性；严格一致性需要更复杂方案，但成本更高。

## 6. GoMQ（发布/消费、QoS、ack/nack、并发 worker）

37. Q：QoS 的作用？
A：控制消费者端未确认消息数量，防止瞬时拉取过多导致内存或处理过载。

38. Q：ack/nack 的区别？
A：ack 表示成功处理并确认，nack 表示失败并可选择重回队列或丢弃。

39. Q：并发 worker 如何实现？
A：消费者端启动固定数量 goroutine，从消息通道拉取并处理。

40. Q：如何解决重复消费？
A：业务侧做幂等性校验（幂等表、去重 key、唯一约束）。

41. Q：如何处理消息堆积？
A：提升消费者并发、扩容分区、增加重试退避与限流。

42. Q：如何保证消息顺序？
A：对同一 key 的消息固定路由到同一分区或单线程消费。

## 7. RPC（GoRPC）

43. Q：RPC 需要关注的核心能力？
A：序列化、连接管理、超时与取消、重试、负载均衡与错误传播。

44. Q：如何设计超时与重试？
A：请求级超时 + 有限次重试 + 只对幂等方法重试。

45. Q：常见负载均衡策略有哪些？
A：轮询、加权轮询、最少连接、基于延迟的选择。

## 8. 性能与压测（GoBench）

46. Q：90 万 QPS 的测量方法？
A：使用微基准/压测工具在单机环境测试纯路由处理逻辑，排除网络 IO 干扰。

47. Q：单请求约 1.2 μs 的含义？
A：表示路由匹配与中间件链执行的平均耗时级别，不包含外部调用成本。

48. Q：GORM 场景 6001 次操作是如何定义的？
A：按固定操作组合（增删改查）执行指定次数，统计成功率与延迟分布。

49. Q：P95/P99 的意义？
A：分别是 95%/99% 请求耗时的上界，用于衡量尾延迟。

50. Q：如何保证压测结果可重复？
A：固定硬件与版本、关闭背景干扰、统一数据集与并发参数。

## 9. 工程化与可测试性

51. Q：如何组织目录结构？
A：按模块拆分，每个模块内部再分核心逻辑、适配层、测试与示例。

52. Q：如何做单测？
A：核心逻辑用纯函数或接口隔离，外部依赖用 mock 或内存实现替代。

53. Q：如何做集成测试？
A：使用 docker 或本地实例启动 Redis/MQ/DB，测试真实链路。

54. Q：如何保证质量？
A：统一 lint、单测、基准测试门槛，结合 CI 流程执行。

## 10. 排障与场景题

55. Q：缓存命中率突然下降怎么排查？
A：先看 key 设计与 TTL 是否异常，再查回源失败率与热点 key 变化。

56. Q：线上 P99 突增怎么排查？
A：先看是否有慢 SQL/外部依赖抖动，再看 GC 与 goroutine 堆积。

57. Q：锁续期失败导致超卖怎么办？
A：立刻降级或临时加大锁 TTL，排查续期线程是否阻塞，增加业务幂等。

58. Q：消息重复消费导致数据异常？
A：补全幂等校验与去重表，增加可观测性记录消息 ID。

59. Q：AutoMigrate 导致线上表锁？
A：只在维护窗口执行，或改为离线迁移脚本，禁用线上自动迁移。

60. Q：路由冲突导致 404？
A：打印路由树与注册顺序，检查路径与方法是否被覆盖或注册遗漏。

## 11. Go 语言八股（结合项目）

61. Q：goroutine 调度模型是什么？
A：G-M-P 模型，通过工作窃取保持多核利用率。

62. Q：map 并发读写会怎样？
A：会触发运行时 panic；需要加锁或使用 sync.Map。

63. Q：channel 的底层是什么？
A：基于环形队列与等待队列，支持同步/异步通信。

64. Q：context.Context 如何取消？
A：通过 cancel function 关闭 Done 通道，调用方需监听 Done。

65. Q：defer 在事务中如何使用？
A：Begin 后 defer Rollback，成功时 Commit 并标记避免重复回滚。

66. Q：如何做性能分析？
A：用 pprof 收集 CPU/内存/阻塞，定位热点函数与资源泄漏。

## 12. 设计取舍与对比

67. Q：和 Gin/Gorm 的差异？
A：功能更精简、偏教学与可控性，保留核心链路可自定义扩展。

68. Q：如果开源，你会补哪些特性？
A：文档、插件系统、更多中间件、更多数据库适配与观测能力。

69. Q：你认为最有挑战的模块？
A：分布式锁与缓存一致性场景，涉及一致性与性能的平衡。

70. Q：最能体现工程化能力的点？
A：模块化边界设计、压测基线与可重复性、错误模型与可观测性。

# GoCache 相关面试问题（补充）

## 1. Singleflight
Q：singleflight 在缓存场景下解决了什么具体问题？
A：合并并发的同 key miss，只让一个请求回源，其他请求等待结果，避免缓存击穿和回源风暴。

Q：如果并发请求同一个 key，回源函数发生 panic，singleflight 如何处理？其他等待请求会怎样？
A：当前实现没有 recover，panic 会导致 wg.Done() 不执行、条目不删除，其它等待请求会一直阻塞。应在 Do 内部 defer recover 并确保 Done 与清理。

## 2. 一致性哈希
Q：节点增加或减少时，一致性哈希如何保证迁移量最小？
A：通过 hash ring 将 key 映射到顺时针最近节点，增删节点只影响邻近区间的 key。

Q：是否实现了虚拟节点（Virtual Nodes）？为什么需要它？
A：实现了，通过 replicas 将一个真实节点映射成多个虚拟节点，提升分布均匀性，减少热点与倾斜。

## 3. 缓存过期策略
Q：TTL 是主动删除还是惰性删除？
A：当前是惰性删除，Get 时检查过期并删除。

Q：惰性删除会不会导致内存无限增长？
A：如果设置了 maxBytes，LRU 会驱逐，内存不会无限增长；若无限制且 key 不再访问，过期项可能滞留。

# 补充：Web/ORM/压测相关面试问答（历史问题汇总）

## 1. Web 框架核心能力与路由
Q：路由器底层数据结构是什么？为什么选它？
A：使用前缀树（Trie/Radix 变体），不是正则或纯哈希表。路径天然是分段结构，Trie 复用公共前缀，匹配复杂度与路径深度相关，性能稳定；比正则更可控，比哈希更灵活（支持参数与通配符）。

Q：动态路径参数（/user/:id、/file/*filepath）如何解析？
A：节点分为静态段、参数段（:id）与通配段（*filepath）。匹配优先静态，其次参数，最后通配。命中参数段记录当前 segment；命中通配段记录剩余路径。

Q：为什么自己写框架而不是直接用 Gin/Echo？你对本质有什么理解？
A：不是替代成熟框架，而是训练对核心链路（路由匹配、中间件模型、上下文、错误恢复、性能优化）的掌控与裁剪能力。Web 框架本质是“路由匹配 + 中间件控制流 + 上下文封装”，其余是工程化扩展。

## 2. 中间件模型与异常控制
Q：中间件是洋葱模型还是链式调用？
A：洋葱模型（Gin 风格）。Context 内维护 handlers 切片和 index，`Next()` 逐个执行，可在 Next 前后做前置/后置逻辑。

Q：鉴权失败如何中断后续链路？
A：通过 `Abort()` 设置中断标志或直接将 index 跳到末尾；`Next()` 检查后直接返回，并可设置 401/403 响应。

Q：panic 如何恢复？recover 写在哪里？
A：最外层 Recovery 中间件里 `defer` 包住 `Next()`；捕获 panic 记录错误并返回 500，防止进程崩溃。
k
## 3. ORM 事务传播与嵌套事务
Q：业务中 `tx := db.Begin()`，Service 内如何使用同一事务？
A：当前实现通过显式传递 `*session.Session`（内部持有 `*sql.Tx`），Service 接收 tx 并使用其方法执行 SQL。

Q：是否通过 context 传递事务对象？
A：当前 `WithContext` 仅用于超时/取消，并不携带事务对象。

Q：嵌套事务如何处理？
A：当前不支持嵌套事务；`Begin()` 在 `tx != nil` 时返回错误。可扩展为计数器/Savepoint。

## 4. 路由并发安全与动态注册
Q：动态注册路由时线程安全吗？
A：当前不安全。Router 使用 map 与切片，未加锁；运行时写会产生数据竞争。

Q：如果要支持动态注册，怎么保证线程安全？
A：可选方案：1）Router 增加 RWMutex，读路径加读锁、注册加写锁；2）Copy-on-Write + atomic.Value，读无锁，写时重建路由结构。

Q：只支持启动时注册，如何防误用？
A：可在 Engine 中加 `started` 标志，在 Run/ServeHTTP 后禁止新增路由，直接返回错误或 panic。

## 5. 90 万 QPS 与 bench 说明
Q：90 万 QPS 是真实 HTTP 吗？
A：不是。90 万来源于 go test 的 micro-benchmark（进程内、无网络），测的是路由匹配+中间件路径的上限。

Q：bench 和真实压测差异为什么大？
A：bench 不走网络栈/系统调用，通常是单 goroutine 循环；真实压测有 TCP/内核栈、并发调度、IO、GC 等开销。

Q：当前仓库的 HTTP 压测工具与配置？
A：使用 GoBench 内置 HTTP 压测，不是 wrk/ab/vegeta。默认 `concurrency=100`、`total=2000`、`url=http://localhost:9999/healthz`。

Q：与 Gin/HttpRouter 的对比数据？
A：当前仓库没有严格同机同工具的对比结果，若需对比应在同环境下跑统一基准。

Q：现在这台机器跑出来是多少？
A：最近一次 `go test -bench BenchmarkEngineHealthz` 的结果约 1292 ns/op，折算 QPS 约 0.77M；不能直接当真实 HTTP QPS。


# 深度问答（按你最新问题整理，含细节）

## 一、Web 框架重构（GoGee）

Q：你在实现路由分组和中间件链时，核心的数据结构设计是怎样的？中间件是如何按顺序执行且支持终止链执行的？
A：
- 路由分组：Engine 维护 `routerGroups []*RouterGroup`，每个 RouterGroup 持有 `prefix` 和 `middlewares`。请求进来时，遍历所有 group，凡是 `path` 有相同前缀的 group，其 middleware 依次拼接到 Context.handles。
- 路由匹配：Router 内部按 Method 维护 `roots map[string]*node`，node 是前缀树（Trie/Radix 变体），根据 path 分段匹配。
- 中间件链：Context 内部有 `handles []HandlerFunc` 和 `index`。`Next()` 会按 `index` 递增顺序执行 handler。中断时 `Abort()` 直接把 `index` 跳到末尾，后续 handler 不再执行。

Q：NoRoute/NoMethod 处理逻辑是如何嵌入到请求生命周期中的？边界场景（路由参数不匹配）如何区分？
A：
- `handle()` 先用 `getRouter(method, path)` 找匹配节点：
  - 找到则追加对应 handler 到链路。
  - 找不到则判断 `pathExists(path)`：如果路径存在但方法不支持，走 NoMethod，设置 `Allow` 头并返回 405；如果路径不存在，走 NoRoute 返回 404。
- 参数不匹配会导致 `getRouter` 返回 nil，此时若同 path 其他方法存在，走 NoMethod；否则走 NoRoute。

Q：模板渲染模块你做了哪些工程化优化？模板缓存、自定义函数、嵌套模板加载失败容错怎么做？
A：
- 模板缓存：Engine 持有 `htmlTemplates *template.Template`，通过 `LoadHTMLGlob` 一次解析并缓存，运行期复用。
- 自定义函数：通过 `SetFuncMap(funcMap)` 注册，`LoadHTMLGlob` 时注入。
- 错误容错：渲染时 `ExecuteTemplate` 写入 buffer，出错直接返回 500；未配置模板时返回统一错误。当前实现没有更复杂的兜底模板机制。

Q：上下文（Context）执行模型如何设计？相比原生 ResponseWriter/Request 做了哪些封装？
A：
- 封装请求参数：`Param()` 获取路由参数，`Query()`/`QueryDefault()` 读 query，`PostForm()` 读表单。
- 响应封装：`JSON()`/`String()`/`Data()`/`HTML()`/`HTMLTemplate()` 统一写响应与 Content-Type。
- 中间件执行：`handles + index + Next()` 实现控制流，`Abort()` 中断。
- 当前版本未实现超时控制（只透传 req 的 context），如果要支持可加入 `Context.Deadline/Done` 与 `WithTimeout`。

## 二、ORM 重构（GoGorm）

Q：事务与 Hook 模块你是如何基于 GORM 扩展的？自定义 Hook 的执行时机控制、嵌套事务回滚怎么处理？
A：
- 不是直接基于 GORM 源码扩展，而是自研轻量 ORM 结构，参考 GORM 思路。
- Hook：定义 `BeforeInsert/AfterInsert/BeforeQuery/AfterQuery` 接口，Insert/Find 调用对应 hook。
- 事务：`Engine.Transaction(fn)` 创建 Session 后 `Begin()`，执行 fn，出错 Rollback，成功 Commit。
- 嵌套事务：当前不支持；`Begin()` 在 `tx != nil` 时直接报错。若要支持可引入 savepoint 或计数器。

Q：AutoMigrate 如何处理“字段新增/删除/类型变更”？字段选择（Select/Omit）如何适配批量与关联查询？
A：
- 当前 AutoMigrate 仅做“表不存在则创建”，不做字段 diff，也不支持删除/类型变更。
- Select：Session 内 `selects []string` 仅在 `Find()` 时生效，限制查询字段；未覆盖 Insert/Update/关联查询。
- 若要完整支持，需要 Schema diff、ALTER TABLE、以及关联/Join 的字段映射与别名处理。

Q：ORM 上下文执行模型如何保证高并发下事务隔离性和查询性能？连接池耗尽怎么处理？
A：
- Session 内持有 `*sql.DB` 与 `*sql.Tx`，事务内操作走 tx，隔离性由数据库级别保证。
- `WithContext(ctx)` 绑定超时/取消，避免请求堆积。
- 连接池由 `sql.DB` 管理；当前代码未暴露 pool 配置，可在 Engine 初始化后由调用方设置 `SetMaxOpenConns/SetMaxIdleConns/SetConnMaxLifetime`。
- 连接池耗尽时表现为请求阻塞或超时，解决方案是调整池大小、缩短事务时间、避免长事务。

## 三、分布式缓存（GoCache）

Q：一致性哈希实现逻辑是什么？虚拟节点数如何选择？
A：
- 实现：`Map` 维护 `keys []int`（hash ring）和 `hashMap map[int]string`（虚拟节点 -> 真实节点）。`Add` 时按 `replicas` 生成虚拟节点；`Get` 用二分查找定位顺时针最近节点。
- 虚拟节点数选择：节点少、负载不均时提升 replicas（如 50-200），提升分布均匀性；代价是内存和构建开销。

Q：singleflight 如何解决缓存击穿？高并发下 goroutine 管理瓶颈如何优化？
A：
- 解决击穿：同 key miss 时只允许一个 goroutine 回源，其它等待结果。
- 当前实现简单（map + WaitGroup），没有超时控制与 panic 保护；高并发下可优化为：
  - 加入 `context` 超时避免等待过久；
  - 引入 panic recover，避免等待者被永久阻塞；
  - 对热点 key 做本地短 TTL 或降级缓存。

Q：TTL 过期策略选择及影响？
A：
- 当前是惰性删除：Get 时检查过期并删除。
- 影响：无后台清理时，过期但不再访问的 key 可能滞留，依赖 LRU 的 maxBytes 限制；若 maxBytes=0，可能长期占用。
- 若要优化，可增加定期清理协程或分层过期队列。

## 四、分布式锁（GoLock）

Q：Redis Lua 脚本实现原子解锁/续期核心逻辑？为什么用 Lua？
A：
- 解锁脚本：`GET key == token` 时 `DEL key`；续期脚本：`GET key == token` 时 `PEXPIRE`。
- Lua 保证“比较 token + 删除/续期”是原子操作，避免误删其他客户端的锁。

Q：自动续期如何设计？如何避免提前释放或永久占用？
A：
- `KeepAlive` 在后台 goroutine 中定时 `Refresh`，默认间隔为 `ttl/3`。
- 避免提前释放：续期间隔要小于 ttl，且 Refresh 校验 token。
- 避免永久占用：业务退出时应主动 Unlock；若进程挂掉，TTL 到期自动释放。

Q：异常场景（服务宕机）如何处理？是否考虑重入/公平锁？
A：
- 宕机：依赖 TTL 自动释放。
- 当前不支持可重入与公平锁；若要支持，可记录 owner + reentrancy count 或引入排队队列。

## 五、消息队列（GoMQ）

Q：QoS（服务质量）如何保证？至少一次/最多一次/恰好一次语义？
A：
- QoS 通过 `channel.Qos(prefetchCount)` 控制未确认消息数量。
- AutoAck=false 且手动 Ack -> 至少一次；AutoAck=true -> 最多一次；恰好一次不保证，需要业务幂等。

Q：ack/nack 机制与重试策略？
A：
- handler 返回错误时 Nack，可根据 `RequeueOnError` 决定是否回队列。
- 当前没有延迟重试/死信队列封装，可通过 RabbitMQ 的 DLX/TTL 机制扩展。

Q：并发 worker 池设计与动态调整？
A：
- 设计：`ConsumeWorkers` 启动固定数量 worker goroutine，从同一 delivery channel 拉取并处理。
- 当前无动态调整策略；如果要支持，可根据 backlog、处理耗时动态调大/缩小 worker 数。

## 六、工程化与性能优化（进阶）

Q：模块化边界如何划分？如何解耦与可插拔？
A：
- Web/ORM/Cache/Lock/MQ/RPC 分模块独立，避免跨模块深耦合。
- 通过接口抽象（如 PeerPicker、TokenGenerator、Dialect）实现可替换。
- 依赖注入：构造函数传入依赖（例如 Redis 客户端、Dialect、PeerPicker）。

Q：GoBench 中 HTTP 微基准与 1.2μs 关系？性能优化做了哪些？
A：
- 两种口径：
  - micro-bench（go test -bench）在进程内测路由匹配，约 1.2μs/op；
  - HTTP 压测（GoBench）在 localhost 测实际请求，约 1.2 万 QPS。
- 当前优化主要来自路由数据结构（Trie）与简化中间件链；没有使用内存池或复杂 GC 优化。

Q：GORM 场景压测尾延迟优化做了哪些？
A：
- 当前以 SQLite 本地压测为主，主要优化点是：减少不必要反射、复用 schema 元信息、缩短事务范围。
- 真正的尾延迟优化通常涉及连接池配置、索引优化、慢 SQL 监控，但这些在当前实现中并未内置。

Q：可测试性如何保证？
A：
- 单测：各模块提供 *_test.go，核心逻辑可在本地运行。
- 集成测试：GoBench 提供统一跑法，支持 HTTP/GORM 模式；Redis/MQ 可通过本地服务验证。
- 性能测试：GoBench 统一输出 JSON 报告，便于对比。

Q：项目可扩展性体现在哪？
A：
- 接口抽象：Dialector、PeerPicker、TokenGenerator、HandlerFunc 等。
- 插件化思路：中间件与 Hook 可扩展；MQ/Cache/Lock 都可通过接口替换实现。

## 七、场景与问题排查（深度）

Q：高并发下出现缓存雪崩，如何定位与优化？
A：
- 定位：看 TTL 分布、缓存命中率、回源 QPS 是否同步抖动。
- 优化：TTL 加随机抖动；热点 key 预热；singleflight 合并回源；限流与降级。

Q：Redis 主从切换导致锁丢失，可能原因与优化？
A：
- 原因：主从延迟导致写未同步，切换后新主丢锁。
- 优化：使用更强一致性存储（如 etcd）、加业务幂等校验、对关键场景缩短锁 TTL 并结合版本校验。

Q：GoMQ 消息堆积，P99 飙升，如何排查？
A：
- 排查维度：生产速度 vs 消费速度；prefetch 是否过大；worker 数是否不足；handler 是否慢或阻塞；MQ 端是否限流。
- 优化：提升并发、拆分队列、加批处理或异步落盘。

Q：Web 框架高并发内存泄漏，如何定位？
A：
- 定位：pprof heap/goroutine，排查 goroutine 泄漏、未关闭 body、巨对象缓存。
- 可能点：中间件启动 goroutine 未退出、context 中缓存大对象、模板缓存无限增长。

Q：落地生产环境需要补哪些能力？
A：
- 监控告警：指标采集、链路追踪、慢请求统计。
- 配置管理：多环境配置与动态刷新。
- 日志体系：结构化日志与采样。
- 降级与限流：请求过载保护。

## 八、设计思路与选型（综合）

Q：为什么自研轻量 Web/ORM 基座？最大挑战是什么？
A：
- 目的：深入理解框架核心链路，做裁剪与定制。
- 挑战：路由与中间件执行模型、ORM 的反射与 schema 映射、以及性能验证。

Q：为何锁/缓存/MQ 基于 Redis 或 RabbitMQ？选型考量？
A：
- Redis：生态成熟、性能高、运维成本低，适合锁/缓存。
- RabbitMQ：协议成熟，支持 QoS/ack/nack，适合任务队列。
- 若要强一致性锁或顺序日志场景，etcd/Kafka 更合适。

Q：工程化重构中遵循哪些 Go 最佳实践？
A：
- 包结构清晰，接口隔离；错误显式返回；并发安全用锁/原子；尽量小函数、可测试。

Q：若支持多集群部署（多 Redis/MQ），如何设计？
A：
- 抽象“集群路由层”，按业务 key/租户选择集群；支持健康检查与自动故障转移；配置中心统一管理。

Q：与 Go-Zero/Kitex/GORM 的差异化？
A：
- 目标是轻量、可读、可裁剪；不追求全功能生态，更适合教学与小型可控场景。