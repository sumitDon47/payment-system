package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	model "github.com/sumitDon47/payment-system/payment-service/models"
)

type Dispatcher struct {
	db           *sql.DB
	writer       *kafka.Writer
	dlqWriter    *kafka.Writer
	dlqTopic     string
	maxRetries   int
	batchSize    int
	pollInterval time.Duration
}

type outboxRecord struct {
	ID       string
	Topic    string
	EventKey string
	Payload  []byte
}

func NewDispatcher(db *sql.DB, broker, defaultTopic, dlqTopic string, maxRetries int) *Dispatcher {
	if defaultTopic == "" {
		defaultTopic = model.TopicPaymentCompleted
	}
	if dlqTopic == "" {
		dlqTopic = model.TopicPaymentDLQ
	}
	if maxRetries <= 0 {
		maxRetries = 5
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        defaultTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		BatchTimeout: 500 * time.Millisecond,
		Async:        false,
	}

	dlqWriter := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        dlqTopic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		BatchTimeout: 500 * time.Millisecond,
		Async:        false,
	}

	return &Dispatcher{
		db:           db,
		writer:       writer,
		dlqWriter:    dlqWriter,
		dlqTopic:     dlqTopic,
		maxRetries:   maxRetries,
		batchSize:    50,
		pollInterval: 2 * time.Second,
	}
}

func (d *Dispatcher) Start(ctx context.Context) {
	log.Println("Outbox dispatcher started")
	defer func() {
		if err := d.writer.Close(); err != nil {
			log.Printf("Outbox writer close error: %v", err)
		}
		if err := d.dlqWriter.Close(); err != nil {
			log.Printf("DLQ writer close error: %v", err)
		}
		log.Println("Outbox dispatcher stopped")
	}()

	if err := d.resetProcessingToPending(ctx); err != nil {
		log.Printf("Outbox reset warning: %v", err)
	}

	for {
		if err := d.dispatchBatch(ctx); err != nil && ctx.Err() == nil {
			log.Printf("Outbox dispatch error: %v", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(d.pollInterval):
		}
	}
}

func (d *Dispatcher) resetProcessingToPending(ctx context.Context) error {
	_, err := d.db.ExecContext(ctx,
		`UPDATE outbox_events SET status = $1 WHERE status = $2`,
		model.OutboxStatusPending,
		model.OutboxStatusProcessing,
	)
	return err
}

func (d *Dispatcher) dispatchBatch(ctx context.Context) error {
	records, err := d.claimBatch(ctx)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}

	for _, rec := range records {
		topic := rec.Topic
		if topic == "" {
			topic = model.TopicPaymentCompleted
		}

		if err := d.writer.WriteMessages(ctx, kafka.Message{
			Key:   []byte(rec.EventKey),
			Value: rec.Payload,
			Time:  time.Now().UTC(),
		}); err != nil {
			if handleErr := d.handlePublishFailure(ctx, rec, err); handleErr != nil {
				log.Printf("Outbox failure handling failed for %s: %v", rec.ID, handleErr)
			}
			continue
		}

		if err := d.markPublished(ctx, rec.ID); err != nil {
			log.Printf("Outbox publish mark failed for %s: %v", rec.ID, err)
		}
	}

	return nil
}

func (d *Dispatcher) claimBatch(ctx context.Context) ([]outboxRecord, error) {
	rows, err := d.db.QueryContext(ctx, `
		WITH picked AS (
			SELECT id, topic, event_key, payload
			FROM outbox_events
			WHERE status = $1
			ORDER BY created_at
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		UPDATE outbox_events AS oe
		SET status = $3
		FROM picked
		WHERE oe.id = picked.id
		RETURNING picked.id, picked.topic, picked.event_key, picked.payload
	`, model.OutboxStatusPending, d.batchSize, model.OutboxStatusProcessing)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]outboxRecord, 0, d.batchSize)
	for rows.Next() {
		var rec outboxRecord
		if err := rows.Scan(&rec.ID, &rec.Topic, &rec.EventKey, &rec.Payload); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (d *Dispatcher) markPublished(ctx context.Context, id string) error {
	_, err := d.db.ExecContext(ctx,
		`UPDATE outbox_events SET status = $1, published_at = NOW(), last_error = NULL WHERE id = $2`,
		model.OutboxStatusPublished,
		id,
	)
	return err
}

func (d *Dispatcher) handlePublishFailure(ctx context.Context, rec outboxRecord, publishErr error) error {
	var status string
	var retryCount int
	errText := publishErr.Error()

	err := d.db.QueryRowContext(ctx, `
		UPDATE outbox_events
		SET
			retry_count = retry_count + 1,
			status = CASE
				WHEN retry_count + 1 >= $2 THEN $3
				ELSE $4
			END,
			last_error = $5
		WHERE id = $1
		RETURNING status, retry_count
	`, rec.ID, d.maxRetries, model.OutboxStatusDead, model.OutboxStatusPending, errText).Scan(&status, &retryCount)
	if err != nil {
		return err
	}

	if status != model.OutboxStatusDead {
		return nil
	}

	log.Printf("Outbox event %s moved to dead after %d retries", rec.ID, retryCount)
	return d.publishToDLQ(ctx, rec, errText, retryCount)
}

func (d *Dispatcher) publishToDLQ(ctx context.Context, rec outboxRecord, reason string, retryCount int) error {
	dlqEnvelope := map[string]any{
		"outbox_id":      rec.ID,
		"event_key":      rec.EventKey,
		"original_topic": rec.Topic,
		"reason":         reason,
		"retry_count":    retryCount,
		"failed_at":      time.Now().UTC().Format(time.RFC3339),
		"payload":        json.RawMessage(rec.Payload),
	}

	dlqPayload, err := json.Marshal(dlqEnvelope)
	if err != nil {
		return fmt.Errorf("marshal dlq payload: %w", err)
	}

	if err := d.dlqWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(rec.EventKey),
		Value: dlqPayload,
		Time:  time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("publish dlq: %w", err)
	}

	return nil
}
