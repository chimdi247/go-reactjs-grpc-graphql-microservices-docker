// Package telemetry centralizes OpenTelemetry setup shared by every
// service (account, catalog, order, graphql): trace export to the
// otel-collector, and a Prometheus-format /metrics + /health HTTP
// endpoint fed by the OTel metrics SDK. Each service just calls Setup()
// once from main() — this is what "auto-instrumentation" means for Go
// (no bytecode agent exists like in Java): gRPC servers/clients and the
// GraphQL HTTP handler are wrapped with otelgrpc/otelhttp, which then
// generate spans for every call automatically, with no manual span code
// anywhere in business logic.
package telemetry

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc/credentials/insecure"
)

// Handle bundles everything a service needs to shut down cleanly and to
// record custom metrics (e.g. account's "total users" business KPI).
type Handle struct {
	Meter    otelmetric.Meter
	Shutdown func(context.Context)
}

// Setup wires up tracing (OTLP -> otel-collector) and metrics (OTel SDK ->
// Prometheus exposition), then starts a background HTTP server exposing
// GET /metrics and GET /health on httpPort. The OTLP endpoint is read from
// OTEL_EXPORTER_OTLP_ENDPOINT (defaults to otel-collector:4317).
func Setup(serviceName string, httpPort int) *Handle {
	ctx := context.Background()

	res, err := sdkresource.New(
		ctx,
		sdkresource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceNamespace("go-grpc-graphql-microservices"),
		),
	)
	if err != nil {
		log.Printf("telemetry: resource init failed: %v", err)
		res = sdkresource.Default()
	}

	shutdownFns := []func(context.Context) error{}

	// ── Traces: OTLP/gRPC to otel-collector ──
	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otlpEndpoint == "" {
		otlpEndpoint = "otel-collector:4317"
	}
	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(otlpEndpoint),
		otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("telemetry: trace exporter init failed: %v", err)
	} else {
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		)
		otel.SetTracerProvider(tp)
		shutdownFns = append(shutdownFns, tp.Shutdown)
	}
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// ── Metrics: OTel SDK -> Prometheus exposition on /metrics ──
	promExporter, err := otelprometheus.New()
	if err != nil {
		log.Fatalf("telemetry: prometheus exporter init failed: %v", err)
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(promExporter),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)
	shutdownFns = append(shutdownFns, mp.Shutdown)

	meter := mp.Meter(serviceName)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	go func() {
		addr := ":" + strconv.Itoa(httpPort)
		log.Printf("telemetry: /metrics and /health listening on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Printf("telemetry: http server stopped: %v", err)
		}
	}()

	return &Handle{
		Meter: meter,
		Shutdown: func(ctx context.Context) {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			for _, fn := range shutdownFns {
				_ = fn(ctx)
			}
		},
	}
}
