package messaging

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// Connect establishes a connection to NATS server
func Connect(natsURL string) (*nats.Conn, error) {
	opts := []nats.Option{
		nats.Name("PIRAMID API"),
		nats.Timeout(30 * time.Second),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(5),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("NATS disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("NATS reconnected to %v", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Printf("NATS connection closed")
		}),
	}

	return nats.Connect(natsURL, opts...)
}

// SetupJetStream initializes JetStream and creates required streams
func SetupJetStream(nc *nats.Conn) (nats.JetStreamContext, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	// Create the suricata events stream
	_, err = js.AddStream(&nats.StreamConfig{
		Name:        "SURICATA_EVENTS",
		Description: "Stream for Suricata eve.json events",
		Subjects:    []string{"suricata.eve", "suricata.alert", "suricata.ssh"},
		Storage:     nats.FileStorage,
		MaxAge:      24 * time.Hour,     // Keep events for 24 hours
		MaxBytes:    1024 * 1024 * 1024, // 1GB max
		Duplicates:  time.Hour,
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		return nil, err
	}

	// Create the ban actions stream
	_, err = js.AddStream(&nats.StreamConfig{
		Name:        "BAN_ACTIONS",
		Description: "Stream for IP ban actions",
		Subjects:    []string{"ban.ip", "unban.ip"},
		Storage:     nats.FileStorage,
		MaxAge:      7 * 24 * time.Hour, // Keep ban actions for 7 days
		Duplicates:  time.Hour,
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		return nil, err
	}

	return js, nil
}

// Publisher handles publishing messages to NATS
type Publisher struct {
	js nats.JetStreamContext
}

// NewPublisher creates a new NATS publisher
func NewPublisher(js nats.JetStreamContext) *Publisher {
	return &Publisher{js: js}
}

// PublishEvent publishes a Suricata event to NATS
func (p *Publisher) PublishEvent(subject string, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = p.js.PublishAsync(subject, data)
	return err
}

// PublishBanAction publishes an IP ban action to NATS
func (p *Publisher) PublishBanAction(ip, reason string) error {
	action := map[string]interface{}{
		"action":    "ban",
		"ip":        ip,
		"reason":    reason,
		"timestamp": time.Now().Unix(),
	}

	data, err := json.Marshal(action)
	if err != nil {
		return err
	}

	_, err = p.js.Publish("ban.ip", data)
	return err
}

// Consumer handles consuming messages from NATS
type Consumer struct {
	js   nats.JetStreamContext
	subs []*nats.Subscription
}

// NewConsumer creates a new NATS consumer
func NewConsumer(js nats.JetStreamContext) *Consumer {
	return &Consumer{js: js}
}

// ConsumeEvents starts consuming Suricata events from NATS
func (c *Consumer) ConsumeEvents(ctx context.Context, handler func([]byte) error) error {
	sub, err := c.js.Subscribe("suricata.*", func(msg *nats.Msg) {
		if err := handler(msg.Data); err != nil {
			log.Printf("Error handling message: %v", err)
			// NACK the message so it can be retried
			msg.Nak()
		} else {
			// ACK the message
			msg.Ack()
		}
	}, nats.Durable("piramid-events-consumer"))

	if err != nil {
		return err
	}

	c.subs = append(c.subs, sub)

	// Wait for context cancellation
	<-ctx.Done()

	// Unsubscribe all subscriptions
	for _, sub := range c.subs {
		sub.Unsubscribe()
	}

	return nil
}

// Close closes all subscriptions
func (c *Consumer) Close() {
	for _, sub := range c.subs {
		sub.Unsubscribe()
	}
}
