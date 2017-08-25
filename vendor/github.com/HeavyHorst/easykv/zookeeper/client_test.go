/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package zookeeper

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/HeavyHorst/easykv/testutils"
	"github.com/tevino/go-zookeeper/zk"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type FilterSuite struct{}

var _ = Suite(&FilterSuite{})

func (s *FilterSuite) TestGetValues(t *C) {
	c, err := New([]string{"127.0.0.1"})
	if err != nil {
		t.Error(err)
	}
	defer c.Close()

	c.client.Create("/premtest", []byte(""), int32(0), zk.WorldACL(zk.PermAll))
	c.client.Create("/premtest/database", []byte(""), int32(0), zk.WorldACL(zk.PermAll))
	c.client.Create("/premtest/database/url", []byte("www.google.de"), int32(0), zk.WorldACL(zk.PermAll))
	c.client.Create("/premtest/database/user", []byte("Boris"), int32(0), zk.WorldACL(zk.PermAll))

	c.client.Create("/remtest", []byte(""), int32(0), zk.WorldACL(zk.PermAll))
	c.client.Create("/remtest/database", []byte(""), int32(0), zk.WorldACL(zk.PermAll))
	c.client.Create("/remtest/database/hosts", []byte(""), int32(0), zk.WorldACL(zk.PermAll))
	c.client.Create("/remtest/database/hosts/192.168.0.1", []byte("test1"), int32(0), zk.WorldACL(zk.PermAll))
	c.client.Create("/remtest/database/hosts/192.168.0.2", []byte("test2"), int32(0), zk.WorldACL(zk.PermAll))

	err = testutils.GetValues(t, c)
	if err != nil {
		t.Error(err)
	}
}

func (s *FilterSuite) TestWatchPrefix(t *C) {
	c, err := New([]string{"127.0.0.1"})
	if err != nil {
		t.Error(err)
	}
	defer c.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		testutils.WatchPrefix(context.Background(), t, c, "/", []string{"/"})
	}()

	time.Sleep(100 * time.Millisecond)
	c.client.Create("/remtest/database/hosts/192.168.0.3", []byte("test3"), int32(0), zk.WorldACL(zk.PermAll))
	c.client.Delete("/remtest/database/hosts/192.168.0.3", int32(0))
	wg.Wait()
}

func (s *FilterSuite) TestWatchPrefixCancel(t *C) {
	c, err := New([]string{"127.0.0.1"})
	if err != nil {
		t.Error(err)
	}
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		testutils.WatchPrefix(ctx, t, c, "/", []string{"/"})
	}()

	cancel()
	wg.Wait()
}
