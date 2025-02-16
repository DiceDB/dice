package otel

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
	meterName = "dicedb-otel"
	otelPort  = "8050"
)

type (
	Otel struct {
		ctx context.Context

		Meter api.Meter
	}
)

var (
	OtelSrv *Otel = NewOtel(context.Background())
)

func NewOtel(ctx context.Context) (dotel *Otel) {
	dotel = &Otel{
		ctx: ctx,
	}
	dotel.setup()
	go dotel.Run()
	return
}

func (dotel *Otel) setup() (err error) {
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

func (dotel *Otel) Close() (err error) {
	// Cleanup tasks
	return
}

func (dotel *Otel) Run() (err error) {
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

func (dotel *Otel) serveMetrics() (err error) {
	http.Handle("/metrics", promhttp.Handler())
	if err = http.ListenAndServe(":"+otelPort, nil); err != nil {
		log.Println("Error starting Otel server", err)
		return
	}
	return
}
