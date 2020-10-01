package telemetry

import (
	"time"

	"github.com/armon/go-metrics"
)

type InmemSink struct {
	Interval int
	Retain   int
}

func (i *InmemSink) Finalize() error {
	return nil
}

func (i *InmemSink) Init() (metrics.MetricSink, error) {
	if i == nil {
		return nil, ErrNilConfig
	}

	sink := metrics.NewInmemSink(time.Duration(i.Interval)*time.Second, time.Duration(i.Retain)*time.Second)
	metrics.DefaultInmemSignal(sink)
	return sink, nil
}
