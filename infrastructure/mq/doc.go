// Package mq 消息队列抽象接口。
// TODO: 实现 RabbitMQ Producer/Consumer 适配此接口。
// TODO: 实现 Kafka Producer/Consumer 适配此接口。
package mq

import "context"

type Producer interface {
	Publish(ctx context.Context, topic string, message []byte) error
	Close() error
}

type Consumer interface {
	Subscribe(ctx context.Context, topic string, handler func([]byte) error) error
	Close() error
}
