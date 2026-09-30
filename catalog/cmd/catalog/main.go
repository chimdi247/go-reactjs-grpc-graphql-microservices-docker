package main

import (
	"context"
	"log"
	"time"

	"github.com/akhilsharma90/go-graphql-microservice/catalog"
	"github.com/akhilsharma90/go-graphql-microservice/pkg/telemetry"
	"github.com/kelseyhightower/envconfig"
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

	tel := telemetry.Setup("catalog", cfg.MetricsPort)
	defer tel.Shutdown(context.Background())

	var r catalog.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		r, err = catalog.NewElasticRepository(cfg.DatabaseURL)
		if err != nil {
			log.Println(err)
		}
		return
	})
	defer r.Close()

	log.Println("Listening on port 8080...")
	s := catalog.NewService(r)
	log.Fatal(catalog.ListenGRPC(s, tel.Meter, 8080))
}
