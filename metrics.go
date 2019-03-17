package konfig

import "github.com/prometheus/client_golang/prometheus"

var (
	// MetricsConfigReload is the label for the prometheus counter for loader reload
	MetricsConfigReload = "konfig_loader_reload"
	// MetricsConfigReloadDuration is the label for the prometheus summary vector for loader reload duration
	MetricsConfigReloadDuration = "konfig_loader_reload_duration"
)

const (
	metricsSuccessLabel = "success"
	metricsFailureLabel = "failure"
)

// LoaderMetrics is the structure holding the promtheus metrics objects
type loaderMetrics struct {
	configReloadSuccess  prometheus.Counter
	configReloadFailure  prometheus.Counter
	configReloadDuration prometheus.Observer
}

func (lw *loaderWatcher) setMetrics() {
	var (
		configReloadCounterVec         = lw.s.metrics[MetricsConfigReload].(*prometheus.CounterVec)
		configReloadDurationSummaryVec = lw.s.metrics[MetricsConfigReloadDuration].(*prometheus.SummaryVec)
	)

	lw.metrics = &loaderMetrics{
		configReloadSuccess: configReloadCounterVec.
			WithLabelValues(
				metricsSuccessLabel,
				lw.s.name,
				lw.Name(),
			),
		configReloadFailure: configReloadCounterVec.
			WithLabelValues(
				metricsFailureLabel,
				lw.s.name,
				lw.Name(),
			),
		configReloadDuration: configReloadDurationSummaryVec.
			WithLabelValues(
				lw.s.name,
				lw.Name(),
			),
	}
}

func (c *S) initMetrics() {
	c.metrics = map[string]prometheus.Collector{
		MetricsConfigReload: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: MetricsConfigReload,
				Help: "Number of config loader reload",
			},
			[]string{"result", "store", "loader"},
		),
		MetricsConfigReloadDuration: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       MetricsConfigReloadDuration,
				Help:       "Histogram for the config reload duration",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"store", "loader"},
		),
	}
}

func (c *S) registerMetrics() error {
	for _, metric := range c.metrics {
		var err = prometheus.Register(metric)
		if err != nil && err != err.(prometheus.AlreadyRegisteredError) {
			return err
		}
	}
	return nil
}
