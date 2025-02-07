# DiceDB Observability with Opentelemetry (Diceotel)

This module implements the observability layer required for DiceDB to emit metrics.
The current implementation integrates Opentelemetry with Prometheus in scraping mode.

## Types of metrics
- Counter, a synchronous instrument that supports non-negative increments
- Asynchronous Counter, an asynchronous instrument which supports non-negative increments
- Histogram, a synchronous instrument that supports arbitrary values that are statistically meaningful, such as histograms, summaries, or percentile
- Synchronous Gauge, a synchronous instrument that supports non-additive values, such as room temperature
- Asynchronous Gauge, an asynchronous instrument that supports non-additive values, such as room temperature
- UpDownCounter, a synchronous instrument that supports increments and decrements, such as the number of active requests
- Asynchronous UpDownCounter, an asynchronous instrument that supports increments and decrements

Different data types like Int64, Float64, etc are supported for each of the above metric types.

## Registering a new metric

The observability module should be available from all other modules. There should always
be one instance of the observability module running.
To use any metric in the repo, first register it with the `diceotel` module [here](/internal/diceotel/metrics.go).
```go
func (dotel *DiceOtel) register() (err error) {
	DiceStartCounter, _ = dotel.Meter.Int64Counter("dicedb_start", api.WithDescription("A counter for the start of the DiceDB server"))
	// Add other metrics here
	return
}
```

Access it anywhere in the repo and call the required functions:
```go
	diceotel.DiceStartCounter.Add(ctx, 1)
```

## Local testing
The repo includes a `prometheus-grafana` folder which contains the required docker files to start the prometheus
and grafana folder locally.
Just run the usual docker-compose commands to use them.

To start the docker containers, run
```bash
docker-compose up -d
```

## Prometheus Integration
Prometheus has multiple modes of collecting data from nodes. In this approach, the module
implements the Prometheus scraping mode where Prometheus fetches data periodically from the
monitored node. The monitored node exposes a HTTP API, currently on port 9090, which the Prometheus
server will scrape data from.

The data is usually in the following format:
```# HELP bar a fun little gauge                                                                                                                                                               
# TYPE bar gauge                                                                                                                                                                            
bar{A="B",C="D",otel_scope_name="go.opentelemetry.io/contrib/examples/prometheus",otel_scope_version=""} 41.469774791457255                                                                 
# HELP baz a histogram with custom buckets and rename
baz_bucket{A="B",C="D",otel_scope_name="go.opentelemetry.io/contrib/examples/prometheus",otel_scope_version="",le="64"} 1 
baz_bucket{A="B",C="D",otel_scope_name="go.opentelemetry.io/contrib/examples/prometheus",otel_scope_version="",le="128"} 1
baz_bucket{A="B",C="D",otel_scope_name="go.opentelemetry.io/contrib/examples/prometheus",otel_scope_version="",le="256"} 2
baz_bucket{A="B",C="D",otel_scope_name="go.opentelemetry.io/contrib/examples/prometheus",otel_scope_version="",le="512"} 2
baz_bucket{A="B",C="D",otel_scope_name="go.opentelemetry.io/contrib/examples/prometheus",otel_scope_version="",le="1024"} 4
baz_bucket{A="B",C="D",otel_scope_name="go.opentelemetry.io/contrib/examples/prometheus",otel_scope_version="",le="2048"} 4
baz_bucket{A="B",C="D",otel_scope_name="go.opentelemetry.io/contrib/examples/prometheus",otel_scope_version="",le="4096"} 4
baz_bucket{A="B",C="D",otel_scope_name="go.opentelemetry.io/contrib/examples/prometheus",otel_scope_version="",le="+Inf"} 4
baz_sum{A="B",C="D",otel_scope_name="go.opentelemetry.io/contrib/examples/prometheus",otel_scope_version=""} 1731                                                                           
baz_count{A="B",C="D",otel_scope_name="go.opentelemetry.io/contrib/examples/prometheus",otel_scope_version=""} 4
# HELP foo_total a simple counter
# TYPE foo_total counter
foo_total{A="B",C="D",otel_scope_name="go.opentelemetry.io/contrib/examples/prometheus",otel_scope_version=""} 5
# HELP go_gc_duration_seconds A summary of the wall-time pause (stop-the-world) duration in garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
go_gc_duration_seconds{quantile="0.75"} 0
go_gc_duration_seconds{quantile="1"} 0
go_gc_duration_seconds_sum 0
go_gc_duration_seconds_count 0
```

