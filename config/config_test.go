/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package config

import (
	"io/ioutil"
	"os"
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
    src = "/tmp/test12345.tmpl"
    dst = "/tmp/test12345.cfg"
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

var expectedTemplates = []*template.Processor{
	&template.Processor{
		Src:  "/tmp/test12345.tmpl",
		Dst:  "/tmp/test12345.cfg",
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

var expected = Configuration{
	LogLevel:  "debug",
	LogFormat: "text",
	Resource: []Resource{
		Resource{
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
	f3, err := ioutil.TempFile("/tmp", "")
	if err != nil {
		t.Error(err)
	}
	defer f3.Close()
	_, err = f3.WriteString(testFile)
	if err != nil {
		t.Error(err)
	}
	s.cfgPath = f3.Name()
}

func (s *FilterSuite) TearDownSuite(t *C) {
	err := os.Remove(s.cfgPath)
	if err != nil {
		t.Log(err)
	}
}

func (s *FilterSuite) TestNewConf(t *C) {
	cfg, err := NewConfiguration(s.cfgPath)
	if err != nil {
		t.Error(err)
	}
	t.Check(cfg, DeepEquals, expected)
}
