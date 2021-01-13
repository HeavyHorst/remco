/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package telemetry

import (
	"errors"

	"github.com/HeavyHorst/remco/pkg/log"
	"github.com/armon/go-metrics"
)

const defaultServiceName = "remco"

// Every sink should implement this interface
type Sink interface {
	Init() (metrics.MetricSink, error)
	Finalize() error
}

// ErrNilConfig is returned if Init is called on a nil Config
var ErrNilConfig = errors.New("config is nil")

// Telemetry represents telemetry configuration
type Telemetry struct {
	Enabled              bool
	ServiceName          string `toml:"service_name"`
	HostName             string
	EnableHostname       bool `toml:"enable_hostname"`
	EnableHostnameLabel  bool `toml:"enable_hostname_label"`
	EnableRuntimeMetrics bool `toml:"enable_runtime_metrics"`
	Sinks                Sinks
}

// Configures metrics and adds FanoutSink with all configured sinks
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
		if t.HostName != "" {
			metricsConf.HostName = t.HostName
		}
		metricsConf.EnableHostname = t.EnableHostname
		metricsConf.EnableRuntimeMetrics = t.EnableRuntimeMetrics
		metricsConf.EnableHostnameLabel = t.EnableHostnameLabel
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

// Finalizes all configured sinks
func (t Telemetry) Stop() error {
	for _, sc := range t.Sinks.GetSinks() {
		err := sc.Finalize()
		if err != nil && err != ErrNilConfig {
			return err
		}
	}

	return nil
}

// Sinks represent sinks configuration
type Sinks struct {
	Inmem      *InmemSink
	Statsd     *StatsdSink
	Statsite   *StatsiteSink
	Prometheus *PrometheusSink
}

// GetSinks returns a slice with all Sinks for easy iteration.
func (c *Sinks) GetSinks() []Sink {
	return []Sink{
		c.Inmem,
		c.Statsd,
		c.Statsite,
		c.Prometheus,
	}
}
