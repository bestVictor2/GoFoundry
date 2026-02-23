# GoLock 数据测试报告

- 数据测试时间：2026-02-23T21:27:03+08:00
- 数据测试方式：go benchmark（锁加解锁基准）
- 场景：TryLockUnlock
- 迭代次数：6324
- 单次耗时：161148 ns/op（0.161148 ms/op）
- 吞吐估算：6205.48 ops/s
- 内存开销：196152 B/op
- 分配次数：806 allocs/op

- 单元测试：ok  	GoLock/lock	0.541s

相关文件：
- `data/golock_benchmark.json`
- `logs/golock_benchmark.log`
- `logs/unit_test_all.log`
