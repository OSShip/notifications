package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/OSShip/notifications/internal/email"
	"github.com/OSShip/utils/kafka"
	"github.com/OSShip/utils/observability"
)

var topics = []string{"listing.events", "enrollment.events", "payment.events", "session.events", "mentor.events"}

type Consumer struct {
	Brokers string
	Sender  *email.Sender
}

func (c *Consumer) Start(ctx context.Context) {
	for _, topic := range topics {
		go c.consumeTopic(ctx, topic)
	}
	slog.Info("notification consumers running", "topics", topics, "brokers", c.Brokers)
}

func (c *Consumer) consumeTopic(ctx context.Context, topic string) {
	reader := kafka.NewReader(c.Brokers, topic, "notifications-group")
	defer reader.Close()
	slog.Info("kafka consumer started", "topic", topic)
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				slog.Info("kafka consumer stopping", "topic", topic)
				return
			}
			slog.Warn("kafka read error", "topic", topic, "err", err)
			continue
		}
		var event kafka.Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			slog.Warn("kafka message parse error", "topic", topic, "err", err)
			continue
		}
		slog.Debug("kafka event received", "topic", topic, "type", event.Type)
		subject, body, ok := email.Render(event.Type, event.Payload)
		if !ok {
			slog.Debug("no template for event", "type", event.Type)
			continue
		}
		to := email.ResolveRecipient(event.Type, event.Payload)
		if to == "" {
			slog.Warn("no recipient for event", "type", event.Type)
			continue
		}
		if err := c.Sender.Send(to, subject, body); err != nil {
			slog.Error("email send failed", "type", event.Type, "to", to, "err", err)
			observability.CaptureError(err, map[string]string{"topic": topic, "event_type": event.Type})
		} else {
			slog.Info("notification sent", "type", event.Type, "to", to, "subject", subject)
		}
	}
}
