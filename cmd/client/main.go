package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Beesy23/peril/internal/gamelogic"
	"github.com/Beesy23/peril/internal/pubsub"
	"github.com/Beesy23/peril/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

var connectStr = "amqp://guest:guest@localhost:5672/"

func main() {
	gob.Register(routing.GameLog{})

	fmt.Println("Starting Peril client...")

	conn, err := amqp.Dial(connectStr)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril game server connected to RabbitMQ!")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("could not get username: %v", err)
	}

	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	gs := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gs.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gs, publishCh),
	)

	if err != nil {
		log.Fatalf("could not subscribe to army moves: %v", err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQueueDurable,
		handlerWar(gs, publishCh),
	)

	if err != nil {
		log.Fatalf("could not subscribe to war declarations: %v", err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gs.GetUsername(),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gs),
	)

	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "move":
			mv, err := gs.CommandMove(words)
			if err != nil {
				fmt.Println(err)
				continue
			}

			err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+mv.Player.Username,
				mv,
			)

			if err != nil {
				fmt.Printf("error: %s\n", err)
				continue
			}
			fmt.Printf("Moved %v to %s\n", len(mv.Units), mv.ToLocation)
		case "spawn":
			err := gs.CommandSpawn(words)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			if len(words) < 2 {
				fmt.Println("spam missing additional argument")
				continue
			}

			n, err := strconv.Atoi(words[1])
			if err != nil {
				fmt.Printf("error: %s\n", err)
				continue
			}
			for i := 0; i < n; i++ {
				msg := gamelogic.GetMaliciousLog()
				err := publishGameLog(
					publishCh,
					gs.GetUsername(),
					msg,
				)
				if err != nil {
					fmt.Printf("error: %s\n", err)
					continue
				}
			}
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Printf("unknown command: %v\n", words[0])
		}
	}
}

func publishGameLog(publishCh *amqp.Channel, username, msg string) error {
	return pubsub.PublishGob(
		publishCh,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+username,
		routing.GameLog{
			Username:    username,
			CurrentTime: time.Now(),
			Message:     msg,
		},
	)
}
