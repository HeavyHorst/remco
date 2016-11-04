/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package main

import (
	"errors"
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/backends/mock"
	"github.com/HeavyHorst/remco/template"

	. "gopkg.in/check.v1"
)

const testFile string = `
log_level = "debug"
log_format = "text"
[[resource]]
  [[resource.template]]
    src = "test12345"
    dst = "test12345"
    checkCmd = ""
    reloadCmd = ""
    mode = "0644"
 [resource.backend]
      [resource.backend.mock]
	    keys = ["/"]
		watch = false
		interval = 1
		onetime = false
`

var expectedTemplates = []*template.ProcessConfig{
	&template.ProcessConfig{
		Src:  "test12345",
		Dst:  "test12345",
		Mode: "0644",
	},
}

var expectedBackend = backends.Config{
	Mock: &mock.Config{
		Backend: template.Backend{
			Watch:    false,
			Keys:     []string{"/"},
			Interval: 1,
			Onetime:  false,
		},
	},
}

var expected = configuration{
	LogLevel:  "debug",
	LogFormat: "text",
	Resource: []resource{
		resource{
			Template: expectedTemplates,
			Backend:  expectedBackend,
		},
	},
}

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type FilterSuite struct {
	cfgPath string
}

var _ = Suite(&FilterSuite{})

func (s *FilterSuite) SetUpSuite(t *C) {
	f, err := ioutil.TempFile("/tmp", "")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()
	_, err = f.WriteString(testFile)
	if err != nil {
		t.Error(err)
	}
	s.cfgPath = f.Name()
}

func (s *FilterSuite) TearDownSuite(t *C) {
	err := os.Remove(s.cfgPath)
	if err != nil {
		t.Log(err)
	}
}

func (s *FilterSuite) TestNewConf(t *C) {
	cfg, err := newConfiguration(s.cfgPath)
	if err != nil {
		t.Error(err)
	}
	t.Check(cfg, DeepEquals, expected)
}

func runTest(cfg configuration, t *C) {
	wait := sync.WaitGroup{}
	stop := make(chan struct{})
	wait.Add(1)
	go func() {
		defer wait.Done()
		cfg.run(stop)
	}()
	close(stop)
	wait.Wait()
}

func (s *FilterSuite) TestRun(t *C) {
	cfg, err := newConfiguration(s.cfgPath)
	if err != nil {
		t.Error(err)
	}
	runTest(cfg, t)
}

// the error should just be logged
func (s *FilterSuite) TestRunWithError(t *C) {
	cfg := expected
	cfg.Resource[0].Backend.Mock.Error = errors.New("test")
	runTest(cfg, t)
}

// should run and exit on its own -- no need to close the stopchan
func (s *FilterSuite) TestRunOnetime(t *C) {
	cfg := expected
	cfg.Resource[0].Backend.Mock.Backend.Onetime = true
	cfg.run(make(chan struct{}))
}
