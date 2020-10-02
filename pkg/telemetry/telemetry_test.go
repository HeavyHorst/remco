/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package telemetry

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type TelemetryTestSuite struct {
	telemetry Telemetry
}

var _ = Suite(&TelemetryTestSuite{})

func (s *TelemetryTestSuite) SetUpSuite(t *C) {
	s.telemetry = Telemetry{
		Enabled:     true,
		ServiceName: "mock",
		Sinks: Sinks{
			Inmem: &InmemSink{
				Interval: 10,
				Retain:   60,
			},
			Statsd: &StatsdSink{
				Addr: "127.0.0.1:7524",
			},
			Statsite: &StatsiteSink{
				Addr: "localhost:7523",
			},
			Prometheus: &PrometheusSink{
				Addr:       "127.0.0.1:2112",
				Expiration: 600,
			},
		},
	}
}

func (s *TelemetryTestSuite) TestInit(t *C) {
	m, err := s.telemetry.Init()
	t.Assert(err, IsNil)
	t.Assert(m, NotNil)
	t.Assert(m.ServiceName, Equals, "mock")

	m.AddSample([]string{"test_sample"}, 42)

	// Wait for the metrics server to start listening
	time.Sleep(1 * time.Second)
	resp, err := http.Get("http://127.0.0.1:2112/metrics")
	t.Assert(err, IsNil)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	t.Assert(err, IsNil)

	t.Assert(string(body), Matches, "(?s).*mock_test_sample.*")
	s.telemetry.Stop()
}

func (s *TelemetryTestSuite) TestStop(t *C) {
	s.telemetry.Init()
	err := s.telemetry.Stop()
	t.Assert(err, IsNil)
	resp, err := http.Get("http://127.0.0.1:2112/metrics")
	t.Assert(err, NotNil)
	t.Assert(resp, IsNil)
}

func (s *TelemetryTestSuite) TestReInit(t *C) {
	s.telemetry.Init()
	err := s.telemetry.Stop()
	t.Assert(err, IsNil)
	s.telemetry.ServiceName = "mock2"
	m2, err := s.telemetry.Init()
	t.Assert(err, IsNil)
	t.Assert(m2, NotNil)
	t.Assert(m2.ServiceName, Equals, "mock2")
	s.telemetry.Stop()
}
