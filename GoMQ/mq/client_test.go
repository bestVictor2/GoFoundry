package mq

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type fakeConn struct {
	ch     *fakeChannel
	closed bool
}

func (f *fakeConn) Channel() (amqpChannel, error) {
	return f.ch, nil
}

func (f *fakeConn) Close() error {
	f.closed = true
	return nil
}

type fakeChannel struct {
	exchangeName string
	exchangeType string
	queueName    string
	bindKey      string
	bindExchange string
	prefetch     int

	pubExchange string
	pubKey      string
	pubPayload  []byte
	pubType     string

	consumeCh chan amqp.Delivery
	closed    bool
}

func (f *fakeChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	f.exchangeName = name
	f.exchangeType = kind
	return nil
}

func (f *fakeChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	f.queueName = name
	return amqp.Queue{Name: name}, nil
}

func (f *fakeChannel) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	f.bindKey = key
	f.bindExchange = exchange
	return nil
}

func (f *fakeChannel) Qos(prefetchCount, prefetchSize int, global bool) error {
	f.prefetch = prefetchCount
	return nil
}

func (f *fakeChannel) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	f.pubExchange = exchange
	f.pubKey = key
	f.pubPayload = append([]byte(nil), msg.Body...)
	f.pubType = msg.ContentType
	return nil
}

func (f *fakeChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	return f.consumeCh, nil
}

func (f *fakeChannel) Close() error {
	f.closed = true
	return nil
}

func TestDialDeclareAndPublish(t *testing.T) {
	cfg := DefaultConfig("amqp://guest:guest@localhost:5672/", "orders.queue")
	cfg.Topology.Exchange = "orders.exchange"
	cfg.Topology.ExchangeType = "topic"
	cfg.Topology.RoutingKey = "orders.created"
	cfg.Consumer.PrefetchCount = 5

	fc := &fakeConn{ch: &fakeChannel{consumeCh: make(chan amqp.Delivery)}}
	oldDial := dialAMQP
	dialAMQP = func(url string) (amqpConnection, error) {
		if url != cfg.URL {
			t.Fatalf("unexpected url: %s", url)
		}
		return fc, nil
	}
	defer func() { dialAMQP = oldDial }()

	client, err := Dial(cfg)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer client.Close()

	if fc.ch.exchangeName != cfg.Topology.Exchange || fc.ch.exchangeType != cfg.Topology.ExchangeType {
		t.Fatalf("exchange declare mismatch")
	}
	if fc.ch.queueName != cfg.Topology.Queue {
		t.Fatalf("queue declare mismatch")
	}
	if fc.ch.bindExchange != cfg.Topology.Exchange || fc.ch.bindKey != cfg.Topology.RoutingKey {
		t.Fatalf("bind mismatch")
	}
	if fc.ch.prefetch != cfg.Consumer.PrefetchCount {
		t.Fatalf("qos mismatch")
	}

	err = client.Publish(context.Background(), Message{Body: []byte("hello")})
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	if fc.ch.pubExchange != cfg.Topology.Exchange || fc.ch.pubKey != cfg.Topology.RoutingKey || string(fc.ch.pubPayload) != "hello" {
		t.Fatalf("publish mismatch")
	}
}

type fakeAcker struct {
	ackCount    int64
	nackCount   int64
	lastRequeue bool
}

func (f *fakeAcker) Ack(tag uint64, multiple bool) error {
	atomic.AddInt64(&f.ackCount, 1)
	return nil
}

func (f *fakeAcker) Nack(tag uint64, multiple bool, requeue bool) error {
	atomic.AddInt64(&f.nackCount, 1)
	f.lastRequeue = requeue
	return nil
}

func (f *fakeAcker) Reject(tag uint64, requeue bool) error {
	atomic.AddInt64(&f.nackCount, 1)
	f.lastRequeue = requeue
	return nil
}

func TestConsumeAckAndNack(t *testing.T) {
	cfg := DefaultConfig("amqp://guest:guest@localhost:5672/", "jobs.queue")
	cfg.Consumer.AutoAck = false
	cfg.Consumer.RequeueOnError = true

	ackOK := &fakeAcker{}
	ackErr := &fakeAcker{}
	fc := &fakeConn{ch: &fakeChannel{consumeCh: make(chan amqp.Delivery, 2)}}
	oldDial := dialAMQP
	dialAMQP = func(url string) (amqpConnection, error) { return fc, nil }
	defer func() { dialAMQP = oldDial }()

	client, err := Dial(cfg)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer client.Close()

	fc.ch.consumeCh <- amqp.Delivery{Body: []byte("ok"), Acknowledger: ackOK, DeliveryTag: 1}
	fc.ch.consumeCh <- amqp.Delivery{Body: []byte("bad"), Acknowledger: ackErr, DeliveryTag: 2}
	close(fc.ch.consumeCh)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = client.Consume(ctx, func(ctx context.Context, msg Delivery) error {
		if string(msg.Body) == "bad" {
			return errors.New("biz error")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("consume failed: %v", err)
	}

	if atomic.LoadInt64(&ackOK.ackCount) != 1 {
		t.Fatalf("expected one ack, got %d", atomic.LoadInt64(&ackOK.ackCount))
	}
	if atomic.LoadInt64(&ackErr.nackCount) != 1 || !ackErr.lastRequeue {
		t.Fatalf("expected nack with requeue")
	}
}

func TestValidateConfig(t *testing.T) {
	cfg := DefaultConfig("", "queue")
	if err := cfg.Validate(); !errors.Is(err, ErrURLRequired) {
		t.Fatalf("expected ErrURLRequired, got %v", err)
	}
	cfg = DefaultConfig("amqp://guest:guest@localhost:5672/", "")
	if err := cfg.Validate(); !errors.Is(err, ErrQueueRequired) {
		t.Fatalf("expected ErrQueueRequired, got %v", err)
	}
}

func TestPublishJSON(t *testing.T) {
	cfg := DefaultConfig("amqp://guest:guest@localhost:5672/", "json.queue")
	fc := &fakeConn{ch: &fakeChannel{consumeCh: make(chan amqp.Delivery)}}
	oldDial := dialAMQP
	dialAMQP = func(url string) (amqpConnection, error) { return fc, nil }
	defer func() { dialAMQP = oldDial }()

	client, err := Dial(cfg)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer client.Close()

	payload := map[string]interface{}{"order_id": 1001, "status": "created"}
	if err = client.PublishJSON(context.Background(), payload, nil); err != nil {
		t.Fatalf("publish json failed: %v", err)
	}
	if fc.ch.pubType != "application/json" {
		t.Fatalf("expected content-type application/json, got %s", fc.ch.pubType)
	}
	var got map[string]interface{}
	if err = json.Unmarshal(fc.ch.pubPayload, &got); err != nil {
		t.Fatalf("unmarshal payload failed: %v", err)
	}
	if got["status"] != "created" {
		t.Fatalf("unexpected json payload: %+v", got)
	}
}

func TestConsumeWorkers(t *testing.T) {
	cfg := DefaultConfig("amqp://guest:guest@localhost:5672/", "worker.queue")
	cfg.Consumer.AutoAck = false
	cfg.Consumer.RequeueOnError = true

	ack1 := &fakeAcker{}
	ack2 := &fakeAcker{}
	ack3 := &fakeAcker{}
	fc := &fakeConn{ch: &fakeChannel{consumeCh: make(chan amqp.Delivery, 3)}}
	oldDial := dialAMQP
	dialAMQP = func(url string) (amqpConnection, error) { return fc, nil }
	defer func() { dialAMQP = oldDial }()

	client, err := Dial(cfg)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer client.Close()

	fc.ch.consumeCh <- amqp.Delivery{Body: []byte("1"), Acknowledger: ack1, DeliveryTag: 1}
	fc.ch.consumeCh <- amqp.Delivery{Body: []byte("2"), Acknowledger: ack2, DeliveryTag: 2}
	fc.ch.consumeCh <- amqp.Delivery{Body: []byte("3"), Acknowledger: ack3, DeliveryTag: 3}
	close(fc.ch.consumeCh)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var handled int64
	err = client.ConsumeWorkers(ctx, 2, func(ctx context.Context, msg Delivery) error {
		atomic.AddInt64(&handled, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("consume workers failed: %v", err)
	}
	if atomic.LoadInt64(&handled) != 3 {
		t.Fatalf("expected handle 3 messages, got %d", atomic.LoadInt64(&handled))
	}
	if atomic.LoadInt64(&ack1.ackCount)+atomic.LoadInt64(&ack2.ackCount)+atomic.LoadInt64(&ack3.ackCount) != 3 {
		t.Fatal("expected all messages acked")
	}
}
