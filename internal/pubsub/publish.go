package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	}
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
	return err
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(val)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{
		ContentType: "application/gob",
		Body:        buf.Bytes(),
	}
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
	return err
}
