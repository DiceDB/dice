package diceotel

import (
	"context"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/otel/exporters/prometheus"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

const (
	meterName    = "dicedb-otel"
	diceOtelPort = "8050"
)

type (
	DiceOtel struct {
		ctx context.Context

		Meter api.Meter

		StartCounter            api.Int64Counter
		CmdLatencyInMsHistogram api.Int64Histogram
	}
)

var (
	DiceotelSrv *DiceOtel = NewDiceOtel(context.Background())
)

func NewDiceOtel(ctx context.Context) (dotel *DiceOtel) {
	dotel = &DiceOtel{
		ctx: ctx,
	}
	dotel.setup()
	go dotel.Run()
	return
}

func (dotel *DiceOtel) setup() (err error) {
	// The exporter embeds a default OpenTelemetry Reader and
	// implements prometheus.Collector, allowing it to be used as
	// both a Reader and Collector.
	exporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}
	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	dotel.Meter = provider.Meter(meterName)

	dotel.register()
	return
}

func (dotel *DiceOtel) Close() (err error) {
	// Cleanup tasks
	return
}

func (dotel *DiceOtel) Run() (err error) {
	log.Println("Starting DiceDB Observability server")

	// Start the prometheus HTTP server and pass the exporter Collector to it
	go dotel.serveMetrics()
	select {
	case <-dotel.ctx.Done():
		dotel.Close()
		return
	}
	log.Println("Exiting dotel.Run()")
	return
}

func (dotel *DiceOtel) serveMetrics() (err error) {
	http.Handle("/metrics", promhttp.Handler())
	if err = http.ListenAndServe(":"+diceOtelPort, nil); err != nil {
		log.Println("Error starting Diceotel server", err)
		return
	}
	return
}
