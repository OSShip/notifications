package main

import (
	"context"
	"net/http"
	"os"
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
	logger := observability.InitLogger("notifications")

	ctx := context.Background()
	sender := email.NewSender(cfg.ResendAPIKey, cfg.FromEmail)
	if cfg.ResendAPIKey == "" {
		logger.Warn("Resend API key not configured, emails will be logged only")
	}
	(&consumer.Consumer{Brokers: cfg.KafkaBrokers, Sender: sender}).Start(ctx)
	logger.Info("kafka consumers started", "brokers", cfg.KafkaBrokers)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(observability.SentryHTTPMiddleware)
	r.Use(observability.SentryRecoverMiddleware("notifications"))
	r.Use(observability.SentryErrorMiddleware("notifications"))
	r.Use(observability.RequestLogMiddleware("notifications"))
	r.Use(observability.PrometheusMiddleware("notifications"))
	r.Get("/health", observability.HealthHandler("notifications"))
	r.Get("/metrics", observability.MetricsHandler().ServeHTTP)

	logger.Info("notifications listening", "port", cfg.Port, "from_email", cfg.FromEmail)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		logger.Error("server failed", "err", err)
		os.Exit(1)
	}
}
