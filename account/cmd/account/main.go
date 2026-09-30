package main

import (
	"context"
	"log"
	"time"

	"github.com/akhilsharma90/go-graphql-microservice/account"
	"github.com/akhilsharma90/go-graphql-microservice/pkg/telemetry"
	"github.com/kelseyhightower/envconfig"
	"go.opentelemetry.io/otel/metric"
	"github.com/tinrab/retry"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
	MetricsPort int    `envconfig:"METRICS_PORT" default:"8081"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	tel := telemetry.Setup("account", cfg.MetricsPort)
	defer tel.Shutdown(context.Background())

	var r account.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		r, err = account.NewPostgresRepository(cfg.DatabaseURL)
		if err != nil {
			log.Println(err)
		}
		return
	})
	defer r.Close()

	s := account.NewService(r)

	// Business KPI: total registered users. An observable (pull-based)
	// gauge — Prometheus's own scrape triggers this callback, so the
	// count is always as fresh as the last scrape with no background
	// polling goroutine needed.
	_, err = tel.Meter.Int64ObservableGauge(
		"account_total_users",
		metric.WithDescription("Total number of registered user accounts"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			count, err := s.GetAccountsCount(ctx)
			if err != nil {
				return err
			}
			o.Observe(int64(count))
			return nil
		}),
	)
	if err != nil {
		log.Printf("telemetry: failed to register account_total_users gauge: %v", err)
	}

	log.Println("Listening on port 8080...")
	log.Fatal(account.ListenGRPC(s, tel.Meter, 8080))
}
