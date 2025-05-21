package pubsub

import (
	"bytes"
	"encoding/gob"
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

	unmarshaller := func(data []byte) (T, error) {
		var target T

		err := json.Unmarshal(data, &target)
		return target, err
	}

	return Subscribe(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
		handler,
		unmarshaller,
	)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
) error {

	unmarshaller := func(data []byte) (T, error) {
		var target T
		buf := bytes.NewReader(data)

		dec := gob.NewDecoder(buf)
		err := dec.Decode(&target)

		return target, err
	}

	return Subscribe(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
		handler,
		unmarshaller,
	)
}

func Subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) Acktype,
	unmarshaller func([]byte) (T, error),
) error {
	ch, queue, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
	)
	if err != nil {
		return fmt.Errorf("could not subscribe to %s: %v", queueName, err)
	}

	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("could not create channel: %v", err)
	}

	go func() {
		defer ch.Close()
		for body := range msgs {
			data, err := unmarshaller(body.Body)
			if err != nil {
				fmt.Printf("could not unmarshal message: %v\n", err)
				continue
			}
			acktype := handler(data)
			switch acktype {
			case Ack:
				body.Ack(false)
			case NackRequeue:
				body.Nack(false, true)
			case NackDiscard:
				body.Nack(false, false)
			}
		}
	}()

	return nil
}
