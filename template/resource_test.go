/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package template

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/HeavyHorst/easyKV/mock"
	"github.com/HeavyHorst/remco/executor"

	. "gopkg.in/check.v1"
)

const (
	tmplString string = "{{ getallkvs() | toPrettyJSON }}"
	tmplFile   string = `[
    {
        "Key": "/some/path/data",
        "Value": "someData"
    }
]`
)

type ResourceSuite struct {
	templateFile string
	backend      Backend
	renderer     *Renderer
	resource     *Resource
}

var _ = Suite(&ResourceSuite{})

func (s *ResourceSuite) SetUpSuite(t *C) {
	// create simple template file
	f, err := ioutil.TempFile("", "template")
	t.Assert(err, IsNil)
	defer f.Close()
	_, err = f.WriteString(tmplString)
	t.Assert(err, IsNil)
	s.templateFile = f.Name()

	// create a backend
	s.backend = Backend{
		Name:     "mock",
		Onetime:  false,
		Watch:    true,
		Prefix:   "/",
		Interval: 1,
		Keys:     []string{"/"},
	}
	s.backend.ReadWatcher, _ = mock.New(nil, map[string]string{"/some/path/data": "someData"})

	// create a renderer
	s.renderer = &Renderer{
		Src:       s.templateFile,
		Dst:       "/tmp/remco-basic-test.conf",
		CheckCmd:  "exit 0",
		ReloadCmd: "exit 0",
	}

	exec := executor.New("", "", "", 0, 0, nil)
	res, err := NewResource([]Backend{s.backend}, []*Renderer{s.renderer}, "test", exec)
	t.Assert(err, IsNil)
	s.resource = res
}

func (s *ResourceSuite) TearDownSuite(t *C) {
	err := os.Remove(s.templateFile)
	t.Check(err, IsNil)
}

func (s *ResourceSuite) TestNewResource(t *C) {
	t.Check(s.resource.backends, HasLen, 1)
	t.Check(s.resource.backends[0].store, NotNil)
	t.Check(s.resource.store, NotNil)
	t.Check(s.resource.logger, NotNil)

	fm := newFuncMap()
	addFuncs(fm, s.resource.store.FuncMap)
	t.Check(s.resource.funcMap, HasLen, len(fm))
	t.Check(s.resource.sources, DeepEquals, []*Renderer{s.renderer})
	t.Check(s.resource.SignalChan, NotNil)
}

func (s *ResourceSuite) TestClose(t *C) {
	s.resource.Close()
}

func (s *ResourceSuite) TestSetVars(t *C) {
	err := s.resource.setVars(s.resource.backends[0])
	t.Check(err, IsNil)
	// the backend trie and the global tree should hold the same values
	t.Check(s.resource.store.GetAllKVs(), DeepEquals, s.resource.backends[0].store.GetAllKVs())
}

func (s *ResourceSuite) TestCreateStageFileAndSync(t *C) {
	_, err := s.resource.createStageFileAndSync()
	t.Check(err, IsNil)
}

func (s *ResourceSuite) TestProcess(t *C) {
	_, err := s.resource.process(s.resource.backends)
	t.Check(err, IsNil)

	data, err := ioutil.ReadFile("/tmp/remco-basic-test.conf")
	t.Assert(err, IsNil)
	t.Check(string(data), Equals, tmplFile)
}

func (s *ResourceSuite) TestMonitor(t *C) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(5 * time.Second)
		cancel()
	}()

	s.resource.Monitor(ctx)
	t.Check(s.resource.Failed, Equals, false)
}

func (s *ResourceSuite) TestMonitorWithBackendError(t *C) {
	s.resource.backends[0].ReadWatcher.(*mock.Client).Err = fmt.Errorf("some error")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(5 * time.Second)
		cancel()
	}()

	s.resource.Monitor(ctx)
	t.Check(s.resource.Failed, Equals, false)
	s.resource.backends[0].ReadWatcher.(*mock.Client).Err = nil
}
