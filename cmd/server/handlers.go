package main

import (
	"fmt"

	"github.com/Beesy23/peril/internal/gamelogic"
	"github.com/Beesy23/peril/internal/pubsub"
	"github.com/Beesy23/peril/internal/routing"
)

func handlerWrite(log routing.GameLog) pubsub.Acktype {
	defer fmt.Print("> ")
	err := gamelogic.WriteLog(log)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return pubsub.NackRequeue
	}
	return pubsub.Ack
}
