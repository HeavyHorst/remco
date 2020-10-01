package telemetry

import "github.com/armon/go-metrics"

type StatsiteSink struct {
	Addr string
}

func (s *StatsiteSink) Init() (metrics.MetricSink, error) {
	if s == nil {
		return nil, ErrNilConfig
	}

	return metrics.NewStatsiteSink(s.Addr)
}

func (s *StatsiteSink) Finalize() error {
	return nil
}
