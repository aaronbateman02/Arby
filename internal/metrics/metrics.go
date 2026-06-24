package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	reg     *prometheus.Registry
	factory promauto.Factory
}

func New() *Metrics {
	reg := prometheus.NewRegistry()
	return &Metrics{
		reg:     reg,
		factory: promauto.With(reg),
	}
}

func (m *Metrics) Counter(name, help string, labels ...string) prometheus.Counter {
	return m.factory.NewCounter(prometheus.CounterOpts{
		Name: name,
		Help: help,
	})
}

func (m *Metrics) CounterVec(name, help string, labelNames []string) *prometheus.CounterVec {
	return m.factory.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	}, labelNames)
}

func (m *Metrics) Histogram(name, help string, buckets []float64) prometheus.Histogram {
	return m.factory.NewHistogram(prometheus.HistogramOpts{
		Name:    name,
		Help:    help,
		Buckets: buckets,
	})
}

func (m *Metrics) Gauge(name, help string) prometheus.Gauge {
	return m.factory.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	})
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.reg, promhttp.HandlerOpts{})
}
