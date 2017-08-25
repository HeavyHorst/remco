/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package env

import (
	"os"
	"testing"

	"github.com/HeavyHorst/easykv/testutils"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type FilterSuite struct{}

var _ = Suite(&FilterSuite{})

func (s *FilterSuite) TestTransform(t *C) {
	dat := transform("/foo/bar/test")
	t.Check(dat, Equals, "FOO_BAR_TEST")
}

func (s *FilterSuite) TestClean(t *C) {
	dat := clean("FOO_BAR_TEST")
	t.Check(dat, Equals, "/foo/bar/test")
}

func (s *FilterSuite) TestWatchPrefix(t *C) {
	c, _ := New()
	testutils.WatchPrefixError(t, c)
}

func (s *FilterSuite) TestGetValues(t *C) {
	//set some env vars
	os.Setenv("PREMTEST_DATABASE_URL", "www.google.de")
	os.Setenv("PREMTEST_DATABASE_USER", "Boris")
	os.Setenv("REMTEST_DATABASE_HOSTS_192.168.0.1", "test1")
	os.Setenv("REMTEST_DATABASE_HOSTS_192.168.0.2", "test2")

	c, _ := New()
	testutils.GetValues(t, c)
}
