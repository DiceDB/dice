package diceotel

import (
	"context"
	"io"
	"strings"

	"fmt"
	"log"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"

	"github.com/stretchr/testify/assert"
)

func publishMetrics(ctx context.Context, dotel *DiceOtel) {

	rng := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec // G404: Use of weak random number generator (math/rand instead of crypto/rand) is ignored as this is not security-sensitive.
	opt := api.WithAttributes(
		attribute.Key("A").String("B"),
		attribute.Key("C").String("D"),
	)

	// This is the equivalent of prometheus.NewCounterVec
	counter, err := dotel.Meter.Float64Counter("foo", api.WithDescription("a simple counter"))
	if err != nil {
		log.Fatal(err)
	}
	counter.Add(ctx, 5, opt)

	gauge, err := dotel.Meter.Float64ObservableGauge("bar", api.WithDescription("a fun little gauge"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = dotel.Meter.RegisterCallback(func(_ context.Context, o api.Observer) error {
		n := -10. + rng.Float64()*(90.) // [-10, 100)
		o.ObserveFloat64(gauge, n, opt)
		return nil
	}, gauge)
	if err != nil {
		log.Fatal(err)
	}
	// This is the equivalent of prometheus.NewHistogramVec
	histogram, err := dotel.Meter.Float64Histogram(
		"baz",
		api.WithDescription("a histogram with custom buckets and rename"),
		api.WithExplicitBucketBoundaries(64, 128, 256, 512, 1024, 2048, 4096),
	)
	if err != nil {
		log.Fatal(err)
	}
	histogram.Record(ctx, 136, opt)
	histogram.Record(ctx, 64, opt)
	histogram.Record(ctx, 701, opt)
	histogram.Record(ctx, 830, opt)

	return

}

func checkMetrics() (content string, err error) {
	httpClient := &http.Client{}
	resp, err := httpClient.Get(fmt.Sprintf("http://localhost:%s/metrics", diceOtelPort))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	content = string(bodyBytes)
	return
}

func TestMetricsPublish(t *testing.T) {
	// Publish some metrics
	publishMetrics(context.Background(), DiceotelSrv)
	checkMetrics()
	content, err := checkMetrics()
	assert.Nil(t, err)
	assert.True(t, strings.Contains(content, "foo_total"))
	assert.True(t, strings.Contains(content, "bar"))
	assert.True(t, strings.Contains(content, "baz_bucket"))

}
