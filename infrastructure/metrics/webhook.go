package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	WebhookRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webhook_requests_total",
			Help: "Total webhook requests by status",
		},
		[]string{"status"},
	)

	WebhookLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "webhook_latency_seconds",
			Help:    "Webhook request latency in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10},
		},
		[]string{"operation"},
	)

	RawEventsByStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "raw_events_by_status",
			Help: "Raw events count by status",
		},
		[]string{"status"},
	)

	SyncJobTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sync_job_total",
			Help: "Sync job execution count",
		},
		[]string{"result"},
	)
)

func Init() {
	prometheus.MustRegister(WebhookRequests)
	prometheus.MustRegister(WebhookLatency)
	prometheus.MustRegister(RawEventsByStatus)
	prometheus.MustRegister(SyncJobTotal)
}
