package otel

import (
	api "go.opentelemetry.io/otel/metric"
)

var (
	StartCounter            api.Int64Counter
	CmdLatencyInMsHistogram api.Int64Histogram
)

func (dotel *Otel) register() (err error) {
	StartCounter, _ = dotel.Meter.Int64Counter("dicedb.start.count", api.WithDescription("A counter for the start of the DiceDB server"))
	CmdLatencyInMsHistogram, _ = dotel.Meter.Int64Histogram("dicedb.command_exec.latency", api.WithUnit("ms"))
	// Add new metrics here
	return
}
