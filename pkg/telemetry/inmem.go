/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package telemetry

import (
	"time"

	"github.com/armon/go-metrics"
)

// InmemSink represents inmem sink configuration
type InmemSink struct {
	Interval int
	Retain   int
}

// Creates a new inmem sink from config and registers DefaultInmemSignal signal (SIGUSR1)
func (i *InmemSink) Init() (metrics.MetricSink, error) {
	if i == nil {
		return nil, ErrNilConfig
	}

	sink := metrics.NewInmemSink(time.Duration(i.Interval)*time.Second, time.Duration(i.Retain)*time.Second)
	metrics.DefaultInmemSignal(sink)
	return sink, nil
}

// Just returns nil
func (i *InmemSink) Finalize() error {
	return nil
}
