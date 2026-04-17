package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sumitDon47/payment-system/notification-service/handler"
	"github.com/sumitDon47/payment-system/notification-service/models"
)

// Consumer holds the Kafka reader and processes messages.
type Consumer struct {
	reader *kafka.Reader
}

// New creates a Consumer connected to the given broker and topic.
//
// groupID — consumer group ID. Kafka remembers which messages this
// group has processed. On restart, it resumes from where it left off.
func New(broker, topic, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: groupID,

		// MinBytes/MaxBytes control how much data Kafka sends per fetch.
		// 10KB min = Kafka batches messages to reduce network round trips.
		// 10MB max = prevents huge memory spikes on large messages.
		MinBytes: 10e3,
		MaxBytes: 10e6,

		// CommitInterval = how often to save our position in the topic.
		// Every second = at worst we reprocess 1 second of events after a crash.
		CommitInterval: time.Second,

		// FirstOffset = if this group has never read this topic, start
		// from the very beginning (not just new messages going forward).
		StartOffset: kafka.FirstOffset,

		// How long to wait for a new message before returning empty.
		MaxWait: 3 * time.Second,
	})

	return &Consumer{reader: reader}
}

// Start begins the consume loop. Blocks until ctx is cancelled.
// Run this in a goroutine from main.go.
//
// The loop:
// 1. Fetch next message from Kafka (blocks until one arrives)
// 2. Deserialize JSON → PaymentEvent struct
// 3. Route to correct handler based on EventType
// 4. Commit the offset so Kafka knows we processed it
// 5. Any error = log and continue, never crash the loop
func (c *Consumer) Start(ctx context.Context) {
	log.Println("🎧 Notification service listening for payment events...")

	for {
		// FetchMessage blocks until a message arrives or ctx is cancelled.
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				// Context cancelled = intentional shutdown
				log.Println("🛑 Consumer stopped")
				return
			}
			log.Printf("❌ Failed to fetch message: %v", err)
			continue
		}

		log.Printf("📨 Received: topic=%s partition=%d offset=%d",
			msg.Topic, msg.Partition, msg.Offset)

		// Process — one bad message must never stop the whole consumer
		if err := c.processMessage(msg); err != nil {
			log.Printf("❌ Process failed at offset=%d: %v", msg.Offset, err)
		}

		// Commit offset — tell Kafka "done with this message"
		// We commit even on processing failure to avoid infinite retry loops.
		// Production systems would send failures to a dead-letter queue first.
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("❌ Failed to commit offset: %v", err)
		}
	}
}

// processMessage deserializes and routes a single Kafka message.
func (c *Consumer) processMessage(msg kafka.Message) error {
	var event models.PaymentEvent

	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to deserialize: %w", err)
	}

	switch event.EventType {
	case models.TopicPaymentCompleted:
		return handler.HandlePaymentCompleted(event)
	case models.TopicPaymentFailed:
		return handler.HandlePaymentFailed(event)
	default:
		log.Printf("⚠️  Unknown event type: %s — skipping", event.EventType)
		return nil
	}
}

// Close shuts down the Kafka reader cleanly.
func (c *Consumer) Close() error {
	log.Println("Closing Kafka consumer...")
	return c.reader.Close()
}
