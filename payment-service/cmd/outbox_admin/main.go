package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sumitDon47/payment-system/payment-service/db"
	model "github.com/sumitDon47/payment-system/payment-service/models"
)

type deadEventRow struct {
	ID         string
	Topic      string
	EventKey   string
	RetryCount int
	LastError  string
	CreatedAt  time.Time
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	db.Connect()

	switch os.Args[1] {
	case "list-dead":
		listDeadCmd := flag.NewFlagSet("list-dead", flag.ExitOnError)
		limit := listDeadCmd.Int("limit", 20, "max number of dead events to list")
		_ = listDeadCmd.Parse(os.Args[2:])

		if err := listDeadEvents(*limit); err != nil {
			log.Fatalf("list-dead failed: %v", err)
		}

	case "replay":
		replayCmd := flag.NewFlagSet("replay", flag.ExitOnError)
		id := replayCmd.String("id", "", "dead outbox event id to replay")
		all := replayCmd.Bool("all", false, "replay all dead outbox events")
		_ = replayCmd.Parse(os.Args[2:])

		if !*all && *id == "" {
			log.Fatal("replay requires --id <event-id> or --all")
		}
		if *all && *id != "" {
			log.Fatal("use either --id or --all, not both")
		}

		if *all {
			updated, err := replayAllDeadEvents()
			if err != nil {
				log.Fatalf("replay --all failed: %v", err)
			}
			log.Printf("Requeued %d dead outbox events", updated)
			return
		}

		updated, err := replayDeadEvent(*id)
		if err != nil {
			log.Fatalf("replay failed: %v", err)
		}
		if updated == 0 {
			log.Printf("No dead event found with id=%s", *id)
			return
		}
		log.Printf("Requeued dead event id=%s", *id)

	default:
		printUsage()
		os.Exit(1)
	}
}

func listDeadEvents(limit int) error {
	if limit <= 0 {
		limit = 20
	}

	rows, err := db.DB.Query(`
		SELECT id, topic, event_key, retry_count, COALESCE(last_error, ''), created_at
		FROM outbox_events
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, model.OutboxStatusDead, limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	events := make([]deadEventRow, 0, limit)
	for rows.Next() {
		var evt deadEventRow
		if err := rows.Scan(&evt.ID, &evt.Topic, &evt.EventKey, &evt.RetryCount, &evt.LastError, &evt.CreatedAt); err != nil {
			return err
		}
		events = append(events, evt)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if len(events) == 0 {
		fmt.Println("No dead outbox events found")
		return nil
	}

	fmt.Printf("Dead outbox events (%d):\n", len(events))
	for _, evt := range events {
		fmt.Printf("- id=%s topic=%s key=%s retries=%d created_at=%s\n", evt.ID, evt.Topic, evt.EventKey, evt.RetryCount, evt.CreatedAt.Format(time.RFC3339))
		if evt.LastError != "" {
			fmt.Printf("  last_error=%s\n", evt.LastError)
		}
	}

	return nil
}

func replayDeadEvent(id string) (int64, error) {
	res, err := db.DB.Exec(`
		UPDATE outbox_events
		SET status = $1,
			retry_count = 0,
			last_error = NULL,
			published_at = NULL
		WHERE id = $2 AND status = $3
	`, model.OutboxStatusPending, id, model.OutboxStatusDead)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

func replayAllDeadEvents() (int64, error) {
	res, err := db.DB.Exec(`
		UPDATE outbox_events
		SET status = $1,
			retry_count = 0,
			last_error = NULL,
			published_at = NULL
		WHERE status = $2
	`, model.OutboxStatusPending, model.OutboxStatusDead)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  outbox-admin list-dead [--limit 20]")
	fmt.Println("  outbox-admin replay --id <event-id>")
	fmt.Println("  outbox-admin replay --all")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run ./cmd/outbox_admin list-dead --limit 10")
	fmt.Println("  go run ./cmd/outbox_admin replay --id 7d6f82c4-9be3-4a7a-84db-5a24b9a40947")
	fmt.Println("  go run ./cmd/outbox_admin replay --all")
}
