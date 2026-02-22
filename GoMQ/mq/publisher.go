package mq

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Message struct {
	Body        []byte
	Headers     map[string]interface{}
	ContentType string
	Mandatory   bool
	Immediate   bool
}

func (c *Client) PublishJSON(ctx context.Context, payload interface{}, headers map[string]interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.Publish(ctx, Message{
		Body:        body,
		Headers:     headers,
		ContentType: "application/json",
	})
}

func (c *Client) Publish(ctx context.Context, msg Message) error {
	if err := c.checkReady(); err != nil {
		return err
	}

	headers := amqp.Table{}
	for k, v := range msg.Headers {
		headers[k] = v
	}
	contentType := msg.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	mode := uint8(amqp.Transient)
	if c.cfg.Topology.Durable {
		mode = amqp.Persistent
	}

	return c.ch.PublishWithContext(
		ctx,
		c.cfg.Topology.Exchange,
		c.cfg.Topology.RoutingKey,
		msg.Mandatory,
		msg.Immediate,
		amqp.Publishing{
			ContentType:  contentType,
			Body:         msg.Body,
			Headers:      headers,
			Timestamp:    time.Now(),
			DeliveryMode: mode,
		},
	)
}
