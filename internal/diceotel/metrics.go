package diceotel

import (
	api "go.opentelemetry.io/otel/metric"
)

var (
	// Counters
	DiceStartCounter api.Int64Counter

	// Histograms
	DiceCmdLatencyInMsHistogram api.Int64Histogram
)

func (dotel *DiceOtel) register() (err error) {
	DiceStartCounter, _ = dotel.Meter.Int64Counter("dicedb.start.count", api.WithDescription("A counter for the start of the DiceDB server"))
	DiceCmdLatencyInMsHistogram, _ = dotel.Meter.Int64Histogram("dicedb.command_exec.latency", api.WithUnit("ms"))
	// Add new metrics here
	return
}
