# GoMQ 数据测试报告

- 数据测试时间：2026-02-23T21:27:03+08:00
- 数据测试方式：go benchmark（消息发布/处理基准）

- 场景：PublishJSON
  迭代次数：2912690，单次耗时：409.7 ns/op（0.00041 ms/op），吞吐估算：2440810.35 ops/s，内存：288 B/op，分配：8 allocs/op
- 场景：ProcessDeliveryAck
  迭代次数：23925066，单次耗时：50.71 ns/op（5.1E-05 ms/op），吞吐估算：19719976.34 ops/s，内存：40 B/op，分配：3 allocs/op

- 单元测试：ok  	GoMQ/mq	0.062s

相关文件：
- `data/gomq_benchmark.json`
- `logs/gomq_benchmark.log`
- `logs/unit_test_all.log`
