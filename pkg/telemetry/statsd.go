/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package telemetry

import "github.com/armon/go-metrics"

// StatsdSink represents statsd sink configuration
type StatsdSink struct {
	Addr string
}

// Creates a new statsd sink from config
func (s *StatsdSink) Init() (metrics.MetricSink, error) {
	if s == nil {
		return nil, ErrNilConfig
	}

	return metrics.NewStatsdSink(s.Addr)
}

// Just returns nil
func (s *StatsdSink) Finalize() error {
	return nil
}
