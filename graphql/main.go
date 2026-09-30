//go:generate go run github.com/99designs/gqlgen
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/handler"
	"github.com/akhilsharma90/go-graphql-microservice/pkg/telemetry"
	"github.com/kelseyhightower/envconfig"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type AppConfig struct {
	AccountURL  string `envconfig:"ACCOUNT_SERVICE_URL"`
	CatalogURL  string `envconfig:"CATALOG_SERVICE_URL"`
	OrderURL    string `envconfig:"ORDER_SERVICE_URL"`
	MetricsPort int    `envconfig:"METRICS_PORT" default:"8081"`
}

func main() {
	var cfg AppConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	tel := telemetry.Setup("graphql-gateway", cfg.MetricsPort)
	defer tel.Shutdown(context.Background())

	s, err := NewGraphQLServer(cfg.AccountURL, cfg.CatalogURL, cfg.OrderURL)
	if err != nil {
		log.Fatal(err)
	}

	// otelhttp traces every HTTP request automatically (no manual span
	// code); HTTPMetricsMiddleware records request-count/latency metrics
	// (see pkg/telemetry); authMiddleware pulls the account ID out of a
	// JWT if present (see auth.go) so resolvers can tell who's calling.
	graphqlHandler := otelhttp.NewHandler(
		telemetry.HTTPMetricsMiddleware(tel.Meter)(authMiddleware(handler.GraphQL(s.ToExecutableSchema()))),
		"graphql",
	)

	mux := http.NewServeMux()
	mux.Handle("/graphql", graphqlHandler)
	mux.Handle("/playground", handler.Playground("akhil", "/graphql"))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// CORS so the React frontend (served from its own origin/port) can
	// call this API directly.
	corsHandler := withCORS(mux)

	log.Println("Listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
