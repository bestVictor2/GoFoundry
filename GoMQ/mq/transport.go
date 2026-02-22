package mq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type amqpConnection interface {
	Channel() (amqpChannel, error)
	Close() error
}

type amqpChannel interface {
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error
	Qos(prefetchCount, prefetchSize int, global bool) error
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Close() error
}

var dialAMQP = func(url string) (amqpConnection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	return &realConnection{conn: conn}, nil
}

type realConnection struct {
	conn *amqp.Connection
}

func (r *realConnection) Channel() (amqpChannel, error) {
	ch, err := r.conn.Channel()
	if err != nil {
		return nil, err
	}
	return &realChannel{ch: ch}, nil
}

func (r *realConnection) Close() error {
	return r.conn.Close()
}

type realChannel struct {
	ch *amqp.Channel
}

func (r *realChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	return r.ch.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, args)
}

func (r *realChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return r.ch.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
}

func (r *realChannel) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	return r.ch.QueueBind(name, key, exchange, noWait, args)
}

func (r *realChannel) Qos(prefetchCount, prefetchSize int, global bool) error {
	return r.ch.Qos(prefetchCount, prefetchSize, global)
}

func (r *realChannel) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return r.ch.PublishWithContext(ctx, exchange, key, mandatory, immediate, msg)
}

func (r *realChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	return r.ch.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
}

func (r *realChannel) Close() error {
	return r.ch.Close()
}
