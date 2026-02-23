# GoCache 数据测试报告

- GET 测试时间：2026-02-23T21:26:04+08:00
- BATCH 测试时间：2026-02-23T21:26:04+08:00
- 数据测试方式：GoBench HTTP 压测

GET 场景：
- 接口：http://localhost:9999/api?key=Tom
- 总请求：3000
- 成功率：100.00%
- QPS：13,649.18
- P95/P99：11.365 / 13.005 ms

BATCH 场景：
- 接口：http://localhost:9999/api/batch?keys=Tom,Sam
- 总请求：3000
- 成功率：100.00%
- QPS：14,174.22
- P95/P99：11.295 / 13.122 ms

- 单元测试：ok  	GoCache/lru	0.044s

相关文件：
- `data/gocache_get_benchmark.json`
- `data/gocache_batch_benchmark.json`
- `logs/unit_test_all.log`
