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
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
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
			acktype := handler(data)
			switch acktype {
			case Ack:
				body.Ack(false)
				fmt.Println("Ack")
			case NackRequeue:
				body.Nack(false, true)
				fmt.Println("NackRequeue")
			case NackDiscard:
				body.Nack(false, false)
				fmt.Println("NackDiscard")
			}
		}
	}()

	return nil
}
