package telemetry

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// These two metric names are deliberately explicit (recorded by our own
// interceptor/middleware, not inferred from a contrib library's internal
// naming) so the Grafana dashboard can query them with confidence:
//   - app_requests_total{method,status}            (counter)
//   - app_request_duration_milliseconds{method}     (histogram, in ms —
//     matches the OTel SDK's default histogram bucket boundaries, which
//     are tuned for millisecond-scale values)
const (
	metricRequestsTotal   = "app_requests_total"
	metricRequestDuration = "app_request_duration_milliseconds"
)

type requestMetrics struct {
	counter   otelmetric.Int64Counter
	histogram otelmetric.Float64Histogram
}

func newRequestMetrics(meter otelmetric.Meter) *requestMetrics {
	counter, _ := meter.Int64Counter(
		metricRequestsTotal,
		otelmetric.WithDescription("Total requests handled (gRPC unary + HTTP)"),
	)
	histogram, _ := meter.Float64Histogram(
		metricRequestDuration,
		otelmetric.WithDescription("Request duration in milliseconds"),
		otelmetric.WithUnit("ms"),
	)
	return &requestMetrics{counter: counter, histogram: histogram}
}

// GRPCUnaryMetricsInterceptor records app_requests_total and
// app_request_duration_milliseconds for every unary gRPC call. Combine
// with grpc.StatsHandler(otelgrpc.NewServerHandler()) for traces —
// interceptors and stats handlers are independent and can both be set.
func GRPCUnaryMetricsInterceptor(meter otelmetric.Meter) grpc.UnaryServerInterceptor {
	m := newRequestMetrics(meter)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		elapsedMs := float64(time.Since(start).Microseconds()) / 1000.0

		st, _ := status.FromError(err)
		m.counter.Add(ctx, 1, otelmetric.WithAttributes(
			attribute.String("method", info.FullMethod),
			attribute.String("status", st.Code().String()),
		))
		m.histogram.Record(ctx, elapsedMs, otelmetric.WithAttributes(
			attribute.String("method", info.FullMethod),
		))

		return resp, err
	}
}

// HTTPMetricsMiddleware does the same for HTTP handlers (the GraphQL
// gateway) — wrap alongside otelhttp.NewHandler (that one's for traces).
func HTTPMetricsMiddleware(meter otelmetric.Meter) func(http.Handler) http.Handler {
	m := newRequestMetrics(meter)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: 200}
			next.ServeHTTP(rec, r)
			elapsedMs := float64(time.Since(start).Microseconds()) / 1000.0

			ctx := r.Context()
			m.counter.Add(ctx, 1, otelmetric.WithAttributes(
				attribute.String("method", r.URL.Path),
				attribute.String("status", strconv.Itoa(rec.status)),
			))
			m.histogram.Record(ctx, elapsedMs, otelmetric.WithAttributes(
				attribute.String("method", r.URL.Path),
			))
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
