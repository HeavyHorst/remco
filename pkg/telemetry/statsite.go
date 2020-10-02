/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package telemetry

import "github.com/armon/go-metrics"

// StatsiteSink represents statsite sink configuration
type StatsiteSink struct {
	Addr string
}

// Creates a new statsite sink from config
func (s *StatsiteSink) Init() (metrics.MetricSink, error) {
	if s == nil {
		return nil, ErrNilConfig
	}

	return metrics.NewStatsiteSink(s.Addr)
}

// Just returns nil
func (s *StatsiteSink) Finalize() error {
	return nil
}
