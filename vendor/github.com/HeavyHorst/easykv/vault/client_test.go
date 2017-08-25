/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package vault

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"testing"

	"github.com/HeavyHorst/easykv/testutils"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type FilterSuite struct{}

var _ = Suite(&FilterSuite{})

var token []byte

func init() {
	var err error
	os.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
	token, err = exec.Command("vault", "read", "-field", "id", "auth/token/lookup-self").Output()
	if err != nil {
		fmt.Println(err)
	}
}

func (s *FilterSuite) TestWatchPrefix(t *C) {
	c, err := New("http://127.0.0.1:8200", "token", WithToken(string(token)))
	if err != nil {
		t.Error(err)
	}

	testutils.WatchPrefixError(t, c)
}

func (s *FilterSuite) TestGetValues(t *C) {
	c, err := New("http://127.0.0.1:8200", "token", WithToken(string(token)))
	if err != nil {
		t.Error(err)
	}

	c.client.Logical().Write("/premtest/database/url", map[string]interface{}{"value": "www.google.de"})
	c.client.Logical().Write("/premtest/database/user", map[string]interface{}{"value": "Boris"})
	c.client.Logical().Write("/remtest/database/hosts", map[string]interface{}{"192.168.0.1": "test1", "192.168.0.2": "test2"})

	testutils.GetValues(t, c)
}

func (s *FilterSuite) TestGetParameterEmptyMap(t *C) {
	var err error
	m := map[string]string{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer panicToError(&err)
		getParameter("test", m)
	}()
	wg.Wait()
	t.Check(err.Error(), Equals, "test is missing from configuration")
}
