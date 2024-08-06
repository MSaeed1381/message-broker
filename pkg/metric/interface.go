package metric

import "github.com/prometheus/client_golang/prometheus"

type Metric interface {
	IncMethodCallCount(method Method, status Status)
	ObserveMethodDuration(method Method, status Status, duration float64)
	IncActiveSubscribers()
	DecActiveSubscribers()
	Serve(reg *prometheus.Registry, prometheusAddress string)
}

type NoImpl struct{}

func (m *NoImpl) IncMethodCallCount(_ Method, _ Status)               {}
func (m *NoImpl) ObserveMethodDuration(_ Method, _ Status, _ float64) {}
func (m *NoImpl) IncActiveSubscribers()                               {}
func (m *NoImpl) DecActiveSubscribers()                               {}
func (m *NoImpl) Serve(_ *prometheus.Registry, _ string)              {}
