package mq

import (
	"context"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Delivery struct {
	Body       []byte
	Headers    map[string]interface{}
	RoutingKey string
	Timestamp  time.Time
}

type HandlerFunc func(ctx context.Context, msg Delivery) error

func (c *Client) Consume(ctx context.Context, handler HandlerFunc) error {
	if err := c.checkReady(); err != nil {
		return err
	}
	if handler == nil {
		return ErrHandlerRequired
	}

	deliveries, err := c.openConsumer()
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case raw, ok := <-deliveries:
			if !ok {
				return nil
			}
			if err = c.processDelivery(ctx, raw, handler); err != nil {
				return err
			}
		}
	}
}

func (c *Client) ConsumeWorkers(ctx context.Context, workers int, handler HandlerFunc) error {
	if workers <= 1 {
		return c.Consume(ctx, handler)
	}
	if err := c.checkReady(); err != nil {
		return err
	}
	if handler == nil {
		return ErrHandlerRequired
	}

	deliveries, err := c.openConsumer()
	if err != nil {
		return err
	}
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 1)
	done := make(chan struct{})
	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for {
			select {
			case <-workerCtx.Done():
				return
			case raw, ok := <-deliveries:
				if !ok {
					return
				}
				if procErr := c.processDelivery(workerCtx, raw, handler); procErr != nil {
					select {
					case errCh <- procErr:
					default:
					}
					cancel()
					return
				}
			}
		}
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker()
	}
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		cancel()
		<-done
		return nil
	case err = <-errCh:
		cancel()
		<-done
		return err
	case <-done:
		return nil
	}
}

func (c *Client) openConsumer() (<-chan amqp.Delivery, error) {
	cc := c.cfg.Consumer
	tc := c.cfg.Topology
	return c.ch.Consume(tc.Queue, cc.Tag, cc.AutoAck, tc.Exclusive, false, tc.NoWait, nil)
}

func (c *Client) processDelivery(ctx context.Context, raw amqp.Delivery, handler HandlerFunc) error {
	cc := c.cfg.Consumer
	err := handler(ctx, Delivery{
		Body:       append([]byte(nil), raw.Body...),
		Headers:    mapFromTable(raw.Headers),
		RoutingKey: raw.RoutingKey,
		Timestamp:  raw.Timestamp,
	})
	if cc.AutoAck {
		return nil
	}
	if err != nil {
		if nackErr := raw.Nack(false, cc.RequeueOnError); nackErr != nil {
			return nackErr
		}
		return nil
	}
	return raw.Ack(false)
}

func mapFromTable(t amqp.Table) map[string]interface{} {
	if len(t) == 0 {
		return nil
	}
	out := make(map[string]interface{}, len(t))
	for k, v := range t {
		out[k] = v
	}
	return out
}
