package server

import (
	"log"

	"github.com/dicedb/dice/config"
)

func init() {
	parser := config.NewConfigParser()
	if err := parser.ParseDefaults(config.DiceConfig); err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
}
