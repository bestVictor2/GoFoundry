package main

import (
	"GoMQ/mq"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var mode string
	var url string
	var exchange string
	var routingKey string
	var queue string
	var message string
	flag.StringVar(&mode, "mode", "publish", "publish or consume")
	flag.StringVar(&url, "url", "amqp://guest:guest@localhost:5672/", "rabbitmq url")
	flag.StringVar(&exchange, "exchange", "", "exchange name, empty means direct queue publish")
	flag.StringVar(&routingKey, "routing-key", "", "routing key, default is queue")
	flag.StringVar(&queue, "queue", "demo.queue", "queue name")
	flag.StringVar(&message, "msg", "hello from GoFoundry", "message body for publish mode")
	flag.Parse()

	cfg := mq.DefaultConfig(url, queue)
	cfg.Topology.Exchange = exchange
	if routingKey != "" {
		cfg.Topology.RoutingKey = routingKey
	}
	client, err := mq.Dial(cfg)
	if err != nil {
		log.Fatalf("dial rabbitmq failed: %v", err)
	}
	defer client.Close()

	switch mode {
	case "publish":
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err = client.PublishJSON(ctx, map[string]interface{}{
			"message": message,
			"sent_at": time.Now().Format(time.RFC3339),
		}, nil); err != nil {
			log.Fatalf("publish json failed: %v", err)
		}
		log.Printf("published json message to queue=%s", queue)
	case "consume":
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		log.Printf("consuming queue=%s, press Ctrl+C to stop", queue)
		err = client.Consume(ctx, func(ctx context.Context, msg mq.Delivery) error {
			log.Printf("received routing_key=%s body=%s", msg.RoutingKey, string(msg.Body))
			return nil
		})
		if err != nil {
			log.Fatalf("consume failed: %v", err)
		}
	default:
		log.Fatalf("unsupported mode: %s", mode)
	}
}
