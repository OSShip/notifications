package config

import "os"

type Config struct {
	Port         string
	KafkaBrokers string
	ResendAPIKey string
	FromEmail    string
}

func Load() Config {
	return Config{
		Port:         env("PORT", "8086"),
		KafkaBrokers: env("KAFKA_BROKERS", "kafka:9092"),
		ResendAPIKey: env("RESEND_API_KEY", ""),
		FromEmail:    env("RESEND_FROM_EMAIL", "noreply@osship.local"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
