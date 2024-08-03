package metric

type Metric interface {
	IncMethodCallCount(method Method, status Status)
	ObserveMethodDuration(method Method, status Status, duration float64)
	IncActiveSubscribers()
	DecActiveSubscribers()
}

type NoImpl struct{}

func (m *NoImpl) IncMethodCallCount(method Method, status Status)                      {}
func (m *NoImpl) ObserveMethodDuration(method Method, status Status, duration float64) {}
func (m *NoImpl) IncActiveSubscribers()                                                {}
func (m *NoImpl) DecActiveSubscribers()                                                {}
