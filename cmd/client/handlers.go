package main

import (
	"fmt"

	"github.com/Beesy23/peril/internal/gamelogic"
	"github.com/Beesy23/peril/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}
