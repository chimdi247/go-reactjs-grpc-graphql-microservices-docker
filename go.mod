module github.com/akhilsharma90/go-graphql-microservice

go 1.22

require (
	github.com/99designs/gqlgen v0.10.2
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/golang/protobuf v1.5.4
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lib/pq v1.10.9
	github.com/prometheus/client_golang v1.20.4
	github.com/segmentio/ksuid v1.0.4
	github.com/tinrab/retry v1.0.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.56.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.56.0
	go.opentelemetry.io/otel v1.31.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.31.0
	go.opentelemetry.io/otel/exporters/prometheus v0.53.0
	go.opentelemetry.io/otel/metric v1.31.0
	go.opentelemetry.io/otel/sdk v1.31.0
	go.opentelemetry.io/otel/sdk/metric v1.31.0
	golang.org/x/crypto v0.28.0
	google.golang.org/grpc v1.67.1
	gopkg.in/olivere/elastic.v5 v5.0.86
)
