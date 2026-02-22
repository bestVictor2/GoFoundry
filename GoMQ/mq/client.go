package mq

type Client struct {
	cfg    Config
	conn   amqpConnection
	ch     amqpChannel
	closed bool
}

func Dial(cfg Config) (*Client, error) {
	cfg = cfg.Normalize()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	conn, err := dialAMQP(cfg.URL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	client := &Client{
		cfg:  cfg,
		conn: conn,
		ch:   ch,
	}

	if err = client.declareTopology(); err != nil {
		_ = client.Close()
		return nil, err
	}
	if err = client.applyQoS(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}

func (c *Client) Close() error {
	if c == nil || c.closed {
		return nil
	}
	c.closed = true

	var err error
	if c.ch != nil {
		if closeErr := c.ch.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if c.conn != nil {
		if closeErr := c.conn.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
}

func (c *Client) declareTopology() error {
	t := c.cfg.Topology
	if t.Exchange != "" {
		if err := c.ch.ExchangeDeclare(t.Exchange, t.ExchangeType, t.Durable, t.AutoDelete, false, t.NoWait, nil); err != nil {
			return err
		}
	}

	if _, err := c.ch.QueueDeclare(t.Queue, t.Durable, t.AutoDelete, t.Exclusive, t.NoWait, nil); err != nil {
		return err
	}

	if t.Exchange != "" {
		if err := c.ch.QueueBind(t.Queue, t.RoutingKey, t.Exchange, t.NoWait, nil); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) applyQoS() error {
	if c.cfg.Consumer.PrefetchCount <= 0 {
		return nil
	}
	return c.ch.Qos(c.cfg.Consumer.PrefetchCount, 0, false)
}

func (c *Client) checkReady() error {
	if c == nil || c.ch == nil || c.closed {
		return ErrClientClosed
	}
	return nil
}
