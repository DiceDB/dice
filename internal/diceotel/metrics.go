package diceotel

import (
	api "go.opentelemetry.io/otel/metric"
)

func (dotel *DiceOtel) register() (err error) {
	dotel.StartCounter, _ = dotel.Meter.Int64Counter("dicedb.start.count", api.WithDescription("A counter for the start of the DiceDB server"))
	dotel.CmdLatencyInMsHistogram, _ = dotel.Meter.Int64Histogram("dicedb.command_exec.latency", api.WithUnit("ms"))
	// Add new metrics here
	return
}
