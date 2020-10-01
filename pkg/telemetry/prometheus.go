package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/HeavyHorst/remco/pkg/log"
	"github.com/armon/go-metrics"
	metricsPrometheus "github.com/armon/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusSink struct {
	Addr       string
	Expiration int

	httpServer     *http.Server
	prometheusSink *metricsPrometheus.PrometheusSink
}

func (p *PrometheusSink) Init() (metrics.MetricSink, error) {
	if p == nil {
		return nil, ErrNilConfig
	}

	handler := http.NewServeMux()
	handler.Handle("/metrics", promhttp.Handler())
	p.httpServer = &http.Server{Addr: p.Addr, Handler: handler}

	go func() {
		err := p.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Error(fmt.Sprintf("error starting prometheus server: %v", err))
		}
	}()

	var err error
	p.prometheusSink, err = metricsPrometheus.NewPrometheusSinkFrom(metricsPrometheus.PrometheusOpts{
		Expiration: time.Duration(p.Expiration) * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return p.prometheusSink, nil
}

func (p *PrometheusSink) Finalize() error {
	if p == nil {
		return ErrNilConfig
	}

	prometheus.Unregister(p.prometheusSink)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.httpServer.Shutdown(ctx)
}
