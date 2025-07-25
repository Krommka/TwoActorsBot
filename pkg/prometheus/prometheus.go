package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	CommandCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bot_commands_total",
			Help: "Count of processed commands",
		},
		[]string{"command", "status"},
	)
	CommandDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bot_command_duration_seconds",
			Help:    "Time taken to process command",
			Buckets: []float64{0.01, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"command"},
	)
	ActiveUsers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "bot_active_users_total",
			Help: "Current number of active users",
		},
	)

	APIFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bot_api_failures_total",
			Help: "Count of failed API calls",
		},
		[]string{"status"},
	)

	MessagesSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bot_messages_sent_total",
			Help: "Count of sent messages",
		},
		[]string{"status"}, // ok, error
	)
	CacheOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_operations_total",
			Help: "Cache operations",
		},
		[]string{"status"},
	)
)

func init() {
	prometheus.MustRegister(
		CommandCounter,
		CommandDuration,
		ActiveUsers,
		APIFailures,
		MessagesSent,
		CacheOperations,
	)
}
