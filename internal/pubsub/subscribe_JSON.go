package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType simpleQueueType,
	handler func(T),
) error {
	ch, _, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
	)
	if err != nil {
		return fmt.Errorf("could not subscribe to %s: %v", queueName, err)
	}

	deliv_ch, err := ch.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("could not create channel: %v", err)
	}

	go func() {
		for body := range deliv_ch {
			var data T
			err := json.Unmarshal(body.Body, &data)
			if err != nil {
				fmt.Printf("could not unmarshal JSON: %v\n", err)
				continue
			}
			handler(data)
			body.Ack(false)
		}
	}()

	return nil
}
