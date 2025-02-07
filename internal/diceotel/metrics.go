package diceotel

import (
	api "go.opentelemetry.io/otel/metric"
)

var (
	// Counters
	DiceStartCounter api.Int64Counter
)

func (dotel *DiceOtel) register() (err error) {
	DiceStartCounter, _ = dotel.Meter.Int64Counter("dicedb_start", api.WithDescription("A counter for the start of the DiceDB server"))
	// Add other metrics here
	return
}
