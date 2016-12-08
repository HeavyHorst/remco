/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package runner

import (
	"os"
	"testing"

	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/config"
	"github.com/HeavyHorst/remco/template"

	. "gopkg.in/check.v1"
)

var exampleTemplates = []*template.Renderer{
	&template.Renderer{
		Src:  "/tmp/test12345.tmpl",
		Dst:  "/tmp/test12345.cfg",
		Mode: "0644",
	},
}

var exampleBackend = backends.Config{
	Mock: &backends.MockConfig{
		Backend: template.Backend{
			Watch:    false,
			Keys:     []string{"/"},
			Interval: 1,
			Onetime:  false,
		},
	},
}

var exampleConfiguration = config.Configuration{
	LogLevel:   "debug",
	LogFormat:  "text",
	IncludeDir: "/tmp/resource.d/",
	PidFile:    "/tmp/remco_test.pid",
	Resource: []config.Resource{
		config.Resource{
			Name:     "test.toml",
			Template: exampleTemplates,
			Backend:  exampleBackend,
		},
	},
}

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type RunnerTestSuite struct {
	runner *Runner
}

var _ = Suite(&RunnerTestSuite{})

func (s *RunnerTestSuite) SetUpSuite(t *C) {
	s.runner = New(exampleConfiguration, nil, make(chan struct{}))
}

func (s *RunnerTestSuite) TestNew(t *C) {
	t.Check(s.runner.stopChan, NotNil)
	t.Check(s.runner.reloadChan, NotNil)
	t.Check(s.runner.signalChans, NotNil)
	t.Check(s.runner.reapLock, IsNil)
	t.Check(s.runner.pidFile, Equals, "/tmp/remco_test.pid")
}

func (s *RunnerTestSuite) TestWritePid(t *C) {
	err := s.runner.writePid(os.Getpid())
	t.Check(err, IsNil)
}

func (s *RunnerTestSuite) TestDeletePid(t *C) {
	err := s.runner.deletePid()
	t.Check(err, IsNil)
}

func (s *RunnerTestSuite) TestSignalChan(t *C) {
	c := make(chan os.Signal, 1)
	s.runner.addSignalChan("id", c)
	s.runner.SendSignal(os.Interrupt)
	t.Check(<-c, Equals, os.Interrupt)

	// channel is full, should not block
	c <- os.Interrupt
	s.runner.SendSignal(os.Interrupt)

	s.runner.removeSignalChan("id")
}

func (s *RunnerTestSuite) TestReload(t *C) {
	new := exampleConfiguration
	new.PidFile = "/tmp/remco_test2.pid"
	s.runner.Reload(new)
}

func (s *RunnerTestSuite) TearDownSuite(t *C) {
	s.runner.Stop()
	t.Check(s.runner.signalChans, HasLen, 0)
}
