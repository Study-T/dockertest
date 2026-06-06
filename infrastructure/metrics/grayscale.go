package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	GrayscaleDecisions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grayscale_decisions_total",
			Help: "Grayscale decisions by mode and result",
		},
		[]string{"mode", "result"},
	)

	GrayscaleWhitelistSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "grayscale_whitelist_size",
			Help: "Number of entries in grayscale whitelist",
		},
	)
)

func InitGrayscale() {
	prometheus.MustRegister(GrayscaleDecisions)
	prometheus.MustRegister(GrayscaleWhitelistSize)
}
