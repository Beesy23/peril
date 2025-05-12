package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	pubsub "github.com/Beesy23/peril/internal/pubsub"
	routing "github.com/Beesy23/peril/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

var connectStr = "amqp://guest:guest@localhost:5672/"

func main() {
	fmt.Println("Starting Peril server...")

	connection, err := amqp.Dial(connectStr)
	if err != nil {
		fmt.Println("Connection Unsuccessful:", err)
		os.Exit(1)
	}
	defer connection.Close()
	fmt.Println("Connection Successful")
	ch, err := connection.Channel()
	if err != nil {
		fmt.Println("Creating connection channel failed")
	}
	err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
	if err != nil {
		fmt.Println("Error publishing JSON")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("\nShutting down Peril server")
}
