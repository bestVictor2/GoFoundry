package mq

import (
	"context"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func BenchmarkPublishJSON(b *testing.B) {
	cfg := DefaultConfig("amqp://guest:guest@localhost:5672/", "bench.queue")
	fc := &fakeConn{ch: &fakeChannel{consumeCh: make(chan amqp.Delivery)}}

	oldDial := dialAMQP
	dialAMQP = func(url string) (amqpConnection, error) {
		return fc, nil
	}
	b.Cleanup(func() { dialAMQP = oldDial })

	client, err := Dial(cfg)
	if err != nil {
		b.Fatalf("dial failed: %v", err)
	}
	defer client.Close()

	payload := map[string]interface{}{
		"order_id": 1001,
		"status":   "created",
	}

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err = client.PublishJSON(ctx, payload, nil); err != nil {
			b.Fatalf("publish json failed: %v", err)
		}
	}
}

func BenchmarkProcessDeliveryAck(b *testing.B) {
	cfg := DefaultConfig("amqp://guest:guest@localhost:5672/", "bench.queue")
	cfg.Consumer.AutoAck = false

	fc := &fakeConn{ch: &fakeChannel{consumeCh: make(chan amqp.Delivery)}}
	oldDial := dialAMQP
	dialAMQP = func(url string) (amqpConnection, error) {
		return fc, nil
	}
	b.Cleanup(func() { dialAMQP = oldDial })

	client, err := Dial(cfg)
	if err != nil {
		b.Fatalf("dial failed: %v", err)
	}
	defer client.Close()

	handler := func(ctx context.Context, msg Delivery) error {
		return nil
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ack := &fakeAcker{}
		raw := amqp.Delivery{
			Body:         []byte("ok"),
			Acknowledger: ack,
			DeliveryTag:  uint64(i + 1),
		}
		if err = client.processDelivery(context.Background(), raw, handler); err != nil {
			b.Fatalf("process delivery failed: %v", err)
		}
	}
}
