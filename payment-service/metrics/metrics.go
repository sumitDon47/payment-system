package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// PaymentCounter tracks total payments processed
	PaymentCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_total",
			Help: "Total number of payment transactions processed",
		},
		[]string{"status", "currency"},
	)

	// PaymentAmount tracks total payment amounts
	PaymentAmount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_amount_histogram",
			Help:    "Payment amount distribution",
			Buckets: prometheus.ExponentialBuckets(10, 2, 10), // 10, 20, 40, 80, ..., 5120
		},
		[]string{"currency"},
	)

	// PaymentDuration tracks payment processing time
	PaymentDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_duration_seconds",
			Help:    "Payment processing time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// TransactionStatus tracks transaction status distribution
	TransactionStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "transaction_status",
			Help: "Current transaction status count",
		},
		[]string{"status"},
	)

	// ErrorCounter tracks errors by type
	ErrorCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_errors_total",
			Help: "Total number of errors in payment processing",
		},
		[]string{"error_type"},
	)

	// DatabaseConnectionPoolSize tracks connection pool stats
	DatabaseConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "database_connections",
			Help: "Database connection pool metrics",
		},
		[]string{"state"}, // open, inuse, idle
	)

	// KafkaProducerLag tracks outbox event publishing lag
	KafkaProducerLag = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kafka_producer_lag_seconds",
			Help:    "Lag between event creation and publishing to Kafka",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"topic"},
	)
)
