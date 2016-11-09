/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package fileutil

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/HeavyHorst/remco/log"
	"github.com/Sirupsen/logrus"
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type TestSuite struct{}

var _ = Suite(&TestSuite{})

func init() {
	log.SetLevel("warning")
}

func testSameFile(t *C, srcTxt, dstTxt string) bool {
	src, err := ioutil.TempFile("", "src")
	if err != nil {
		t.Error(err.Error())
	}
	defer os.Remove(src.Name())
	dst, err := ioutil.TempFile("", "dest")
	if err != nil {
		t.Error(err.Error())
	}
	defer os.Remove(dst.Name())

	_, err = src.WriteString(srcTxt)
	if err != nil {
		t.Error(err.Error())
	}
	_, err = dst.WriteString(dstTxt)
	if err != nil {
		t.Error(err.Error())
	}
	status, err := SameFile(src.Name(), dst.Name(), logrus.WithFields(logrus.Fields{}))
	if err != nil {
		t.Error(err.Error())
	}
	return status
}

func (s *TestSuite) TestSameFileTrue(t *C) {
	status := testSameFile(t, "bla", "bla")
	t.Check(status, Equals, true)
}

func (s *TestSuite) TestSameFileFalse(t *C) {
	status := testSameFile(t, "bla", "bla2")
	t.Check(status, Equals, false)
}
