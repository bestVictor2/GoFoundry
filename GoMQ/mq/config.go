package mq

type Topology struct {
	Exchange     string
	ExchangeType string
	Queue        string
	RoutingKey   string
	Durable      bool
	AutoDelete   bool
	Exclusive    bool
	NoWait       bool
}

type ConsumerConfig struct {
	Tag            string
	AutoAck        bool
	PrefetchCount  int
	RequeueOnError bool
}

type Config struct {
	URL      string
	Topology Topology
	Consumer ConsumerConfig
}

func DefaultConfig(url, queue string) Config {
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

func (c Config) Validate() error {
	if c.URL == "" {
		return ErrURLRequired
	}
	if c.Topology.Queue == "" {
		return ErrQueueRequired
	}
	return nil
}
