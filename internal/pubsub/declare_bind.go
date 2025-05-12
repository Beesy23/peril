package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type simpleQueueType int

const (
	SimpleQueueDurable simpleQueueType = iota
	SimpleQueueTransient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType simpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not create channel: %v", err)
	}

	queue, err := ch.QueueDeclare(
		queueName,
		simpleQueueType == SimpleQueueDurable,
		simpleQueueType != SimpleQueueDurable,
		simpleQueueType != SimpleQueueDurable,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not create channel: %v", err)
	}

	err = ch.QueueBind(
		queue.Name,
		key,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not create channel: %v", err)
	}
	return ch, queue, nil
}
