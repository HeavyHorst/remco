/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package consul

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/HeavyHorst/easykv/testutils"
	"github.com/hashicorp/consul/api"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type FilterSuite struct{}

var _ = Suite(&FilterSuite{})

func (s *FilterSuite) TestGetValues(t *C) {
	c, err := New([]string{"localhost:8500"}, WithScheme("http"))
	if err != nil {
		t.Error(err)
	}
	defer c.Close()

	c.client.Put(&api.KVPair{Key: "premtest/database/url", Value: []byte("www.google.de")}, nil)
	c.client.Put(&api.KVPair{Key: "premtest/database/user", Value: []byte("Boris")}, nil)
	c.client.Put(&api.KVPair{Key: "remtest/database/hosts/192.168.0.1", Value: []byte("test1")}, nil)
	c.client.Put(&api.KVPair{Key: "remtest/database/hosts/192.168.0.2", Value: []byte("test2")}, nil)

	testutils.GetValues(t, c)
}

func (s *FilterSuite) TestWatchPrefix(t *C) {
	c, err := New([]string{"localhost:8500"}, WithScheme("http"))
	if err != nil {
		t.Error(err)
	}
	defer c.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		testutils.WatchPrefix(context.Background(), t, c, "/", []string{})
	}()

	time.Sleep(100 * time.Millisecond)
	c.client.Put(&api.KVPair{Key: "remtest/database/hosts/192.168.0.3", Value: []byte("test3")}, nil)
	c.client.Delete("remtest/database/hosts/192.168.0.3", nil)
	wg.Wait()
}

func (s *FilterSuite) TestWatchPrefixCancel(t *C) {
	c, err := New([]string{"localhost:8500"}, WithScheme("http"))
	if err != nil {
		t.Error(err)
	}
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		testutils.WatchPrefix(ctx, t, c, "/", []string{})
	}()

	cancel()
	wg.Wait()
}
