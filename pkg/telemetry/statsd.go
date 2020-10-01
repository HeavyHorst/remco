package telemetry

import "github.com/armon/go-metrics"

type StatsdSink struct {
	Addr string
}

func (s *StatsdSink) Init() (metrics.MetricSink, error) {
	if s == nil {
		return nil, ErrNilConfig
	}

	return metrics.NewStatsdSink(s.Addr)
}

func (s *StatsdSink) Finalize() error {
	return nil
}
