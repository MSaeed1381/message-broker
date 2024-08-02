package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

type PrometheusController struct {
	methodCounter     *prometheus.CounterVec
	methodDuration    *prometheus.SummaryVec
	activeSubscribers *prometheus.GaugeVec
}

func NewPrometheusController(reg prometheus.Registerer) *PrometheusController {
	m := &PrometheusController{
		methodCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "message_broker_method_count",
				Help: "Count of failed/successful RPC calls",
			},
			[]string{"method", "status"},
		),

		methodDuration: promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       "rpc_method_duration_seconds",
				Help:       "Latency of each call in seconds",
				Objectives: map[float64]float64{0.5: 0.05, 0.95: 0.005, 0.99: 0.001},
			}, []string{"method", "status"}),

		activeSubscribers: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "message_broker_active_subscribers",
				Help: "Total active subscriptions",
			},
			[]string{},
		),
	}

	reg.MustRegister(m.methodCounter, m.methodDuration, m.activeSubscribers)
	return m
}

func (p *PrometheusController) IncMethodCallCount(method Method, status Status) {
	p.methodCounter.WithLabelValues(MethodToStr(method), StatusToStr(status)).Inc()
}

func (p *PrometheusController) ObserveMethodDuration(method Method, status Status, duration time.Duration) {
	p.methodDuration.WithLabelValues(MethodToStr(method), StatusToStr(status)).Observe(float64(duration))
}

func (p *PrometheusController) IncActiveSubscribers() {
	p.activeSubscribers.WithLabelValues().Inc()
}

func (p *PrometheusController) DecActiveSubscribers() {
	p.activeSubscribers.WithLabelValues().Dec()
}

func (p *PrometheusController) Serve(reg *prometheus.Registry) {
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	err := http.ListenAndServe(":5555", nil)
	if err != nil {
		panic(err)
	}
}
