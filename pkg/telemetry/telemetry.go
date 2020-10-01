package telemetry

import (
	"errors"

	"github.com/HeavyHorst/remco/pkg/log"
	"github.com/armon/go-metrics"
)

const defaultServiceName = "remco"

type Sink interface {
	Init() (metrics.MetricSink, error)
	Finalize() error
}

var ErrNilConfig = errors.New("config is nil")

type Telemetry struct {
	Enabled     bool
	ServiceName string `toml:"service_name"`
	Sinks       Sinks
}

func (t Telemetry) Init() (*metrics.Metrics, error) {
	var (
		m   *metrics.Metrics
		err error
	)
	if t.Enabled {
		log.Info("enabling telemetry")
		serviceName := defaultServiceName
		if t.ServiceName != "" {
			serviceName = t.ServiceName
		}
		metricsConf := metrics.DefaultConfig(serviceName)
		var sinks metrics.FanoutSink
		for _, sc := range t.Sinks.GetSinks() {
			sink, err := sc.Init()
			if err == nil {
				sinks = append(sinks, sink)
			} else if err != ErrNilConfig {
				return nil, err
			}
		}
		m, err = metrics.NewGlobal(metricsConf, sinks)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

func (t Telemetry) Stop() error {
	for _, sc := range t.Sinks.GetSinks() {
		err := sc.Finalize()
		if err != nil && err != ErrNilConfig {
			return err
		}
	}

	return nil
}

type Sinks struct {
	Inmem      *InmemSink
	Statsd     *StatsdSink
	Statsite   *StatsiteSink
	Prometheus *PrometheusSink
}

func (c *Sinks) GetSinks() []Sink {
	return []Sink{
		c.Inmem,
		c.Statsd,
		c.Statsite,
		c.Prometheus,
	}
}
