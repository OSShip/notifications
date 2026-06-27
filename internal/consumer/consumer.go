package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/OSShip/notifications/internal/email"
	"github.com/OSShip/utils/kafka"
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
}

func (c *Consumer) consumeTopic(ctx context.Context, topic string) {
	reader := kafka.NewReader(c.Brokers, topic, "notifications-group")
	defer reader.Close()
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("kafka read error [%s]: %v", topic, err)
			continue
		}
		var event kafka.Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			continue
		}
		subject, body, ok := email.Render(event.Type, event.Payload)
		if !ok {
			continue
		}
		to := email.ResolveRecipient(event.Type, event.Payload)
		if to == "" {
			log.Printf("no recipient for %s, skipping", event.Type)
			continue
		}
		if err := c.Sender.Send(to, subject, body); err != nil {
			log.Printf("email error: %v", err)
		} else {
			log.Printf("sent notification: %s -> %s", event.Type, to)
		}
	}
}
