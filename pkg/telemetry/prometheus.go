/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

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

// PrometheusSink represents prometheus sink and prometheus stats endpoint configuration
type PrometheusSink struct {
	Addr       string
	Expiration int

	httpServer     *http.Server
	prometheusSink *metricsPrometheus.PrometheusSink
}

// Creates a new prometheus sink from config and starts a goroutine with prometheus stats endpoint
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
			log.Error(fmt.Sprintf("error starting prometheus stats endpoint: %v", err))
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

// Unregisters prometheus sink and stops prometheus stats endpoint
func (p *PrometheusSink) Finalize() error {
	if p == nil {
		return ErrNilConfig
	}

	prometheus.Unregister(p.prometheusSink)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.httpServer.Shutdown(ctx)
}
