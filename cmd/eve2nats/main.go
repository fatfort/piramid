package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"piramid/internal/algo"
	"piramid/internal/config"
	"piramid/internal/database"
	"piramid/internal/messaging"
)

func main() {
	log.Println("Starting eve2nats bridge...")

	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize GeoIP database
	geoipDB, err := algo.NewGeoIPDB(cfg.GeoIPDBPath)
	if err != nil {
		log.Printf("Warning: Failed to load GeoIP database: %v", err)
		geoipDB = nil
	}
	defer func() {
		if geoipDB != nil {
			geoipDB.Close()
		}
	}()

	// Initialize parser
	parser := algo.NewParser(geoipDB)

	// Connect to NATS
	natsConn, err := messaging.Connect(cfg.NATSUrl)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsConn.Close()

	// Initialize JetStream
	js, err := messaging.SetupJetStream(natsConn)
	if err != nil {
		log.Fatalf("Failed to setup JetStream: %v", err)
	}

	// Create publisher
	publisher := messaging.NewPublisher(js)

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received shutdown signal...")
		cancel()
	}()

	// Process stdin (Suricata eve.json output)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			log.Println("Shutting down...")
			return
		default:
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			// Parse the event
			// Default to tenant ID 1 for now (production-tenant)
			tenantID := uint(1)
			parsedEvent, err := parser.ParseEvent(line, tenantID)
			if err != nil {
				log.Printf("Failed to parse event: %v", err)
				continue
			}

			// Store in database
			dbEvent := &database.Event{
				TenantID:   parsedEvent.TenantID,
				Timestamp:  parsedEvent.Timestamp,
				EventType:  parsedEvent.EventType,
				SrcIP:      parsedEvent.SrcIP,
				SrcPort:    parsedEvent.SrcPort,
				DestIP:     parsedEvent.DestIP,
				DestPort:   parsedEvent.DestPort,
				Protocol:   parsedEvent.Protocol,
				Signature:  parsedEvent.Signature,
				Severity:   parsedEvent.Severity,
				Category:   parsedEvent.Category,
				Action:     parsedEvent.Action,
				Country:    parsedEvent.Country,
				City:       parsedEvent.City,
				Latitude:   parsedEvent.Latitude,
				Longitude:  parsedEvent.Longitude,
				RawPayload: parsedEvent.RawPayload,
			}

			if err := db.Create(dbEvent).Error; err != nil {
				log.Printf("Failed to store event in database: %v", err)
			}

			// Determine NATS subject based on event type
			subject := "suricata.eve"
			if parsedEvent.EventType != "" {
				subject = "suricata." + parsedEvent.EventType
			}

			// Publish to NATS
			if err := publisher.PublishEvent(subject, parsedEvent); err != nil {
				log.Printf("Failed to publish event to NATS: %v", err)
			}

			// Log processed event (for debugging)
			log.Printf("Processed %s event from %s -> %s",
				parsedEvent.EventType, parsedEvent.SrcIP, parsedEvent.DestIP)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from stdin: %v", err)
	}

	log.Println("eve2nats bridge stopped")
}
