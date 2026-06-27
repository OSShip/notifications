package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/OSShip/notifications/internal/config"
	"github.com/OSShip/notifications/internal/consumer"
	"github.com/OSShip/notifications/internal/email"
	"github.com/OSShip/utils/observability"
)

func main() {
	cfg := config.Load()
	observability.InitSentry("notifications")
	defer observability.FlushSentry(2 * time.Second)

	ctx := context.Background()
	sender := email.NewSender(cfg.ResendAPIKey, cfg.FromEmail)
	(&consumer.Consumer{Brokers: cfg.KafkaBrokers, Sender: sender}).Start(ctx)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(observability.SentryRecoverMiddleware("notifications"))
	r.Use(observability.SentryErrorMiddleware("notifications"))
	r.Use(observability.PrometheusMiddleware("notifications"))
	r.Get("/health", observability.HealthHandler("notifications"))
	r.Get("/metrics", observability.MetricsHandler().ServeHTTP)

	log.Printf("notifications listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
