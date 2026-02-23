package mq

type Topology struct {
	Exchange     string
	ExchangeType string
	Queue        string
	RoutingKey   string // exchange - > queue
	Durable      bool   // 持久化参数
	AutoDelete   bool   // no consumer
	Exclusive    bool   // 是否只属于当前连接
	NoWait       bool   // 是否异步声明
}

type ConsumerConfig struct {
	Tag            string
	AutoAck        bool
	PrefetchCount  int // 最多持有多少条未 ack 消息
	RequeueOnError bool
}

type Config struct {
	URL      string
	Topology Topology
	Consumer ConsumerConfig
}

func DefaultConfig(url, queue string) Config { // 默认可靠队列，串行消费，失败重试
	return Config{
		URL: url,
		Topology: Topology{
			Queue:        queue,
			RoutingKey:   queue,
			ExchangeType: "direct",
			Durable:      true,
		},
		Consumer: ConsumerConfig{
			AutoAck:        false,
			PrefetchCount:  1,
			RequeueOnError: true,
		},
	}
}

func (c Config) Normalize() Config {
	if c.Topology.RoutingKey == "" {
		c.Topology.RoutingKey = c.Topology.Queue
	}
	if c.Topology.Exchange != "" && c.Topology.ExchangeType == "" {
		c.Topology.ExchangeType = "direct"
	}
	return c
}

func (c Config) Validate() error { // 参数校验
	if c.URL == "" {
		return ErrURLRequired
	}
	if c.Topology.Queue == "" {
		return ErrQueueRequired
	}
	return nil
}
