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

	"github.com/HeavyHorst/remco/pkg/log"
	"github.com/Sirupsen/logrus"
	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
	file          *os.File
	sameFile      *os.File
	differentHash *os.File

	replaceFile  *os.File
	replaceFile1 *os.File
}

var _ = Suite(&TestSuite{})

func init() {
	log.SetLevel("debug")
}

func writeFile(t *C, file *os.File, value string) {
	_, err := file.WriteString(value)
	if err != nil {
		t.Error(err.Error())
	}
}

func createTempFile(t *C, name string) *os.File {
	tempFile, err := ioutil.TempFile("", name)
	if err != nil {
		t.Error(err.Error())
	}
	return tempFile
}

func (s *TestSuite) SetUpSuite(t *C) {
	src := createTempFile(t, "src")
	defer src.Close()

	dst := createTempFile(t, "dest")
	defer dst.Close()

	differentHash := createTempFile(t, "hash")
	defer differentHash.Close()

	replaceFile := createTempFile(t, "replace")
	defer replaceFile.Close()
	replaceFile1 := createTempFile(t, "replace1")
	defer replaceFile1.Close()

	writeFile(t, src, "bla")
	writeFile(t, dst, "bla")
	writeFile(t, replaceFile, "bla")
	writeFile(t, differentHash, "mmmh lecker Gurkensalat!")

	s.file = src
	s.sameFile = dst
	s.differentHash = differentHash
	s.replaceFile = replaceFile
	s.replaceFile1 = replaceFile1
}

func (s *TestSuite) TearDownSuite(t *C) {
	os.Remove(s.file.Name())
	os.Remove(s.sameFile.Name())
	os.Remove(s.differentHash.Name())
	os.Remove(s.replaceFile.Name())
	os.Remove(s.replaceFile1.Name())
}

func (s *TestSuite) TestSameFileTrue(t *C) {
	status, err := SameFile(s.file.Name(), s.sameFile.Name(), logrus.NewEntry(logrus.StandardLogger()))
	if err != nil {
		t.Error(err.Error())
	}
	t.Check(status, Equals, true)
}

func (s *TestSuite) TestSameFileFalse(t *C) {
	status, err := SameFile(s.file.Name(), s.differentHash.Name(), logrus.NewEntry(logrus.StandardLogger()))
	if err != nil {
		t.Error(err.Error())
	}
	t.Check(status, Equals, false)
}

func (s *TestSuite) TestReplaceFile(t *C) {
	fileStat, err := stat(s.file.Name())
	if err != nil {
		t.Error(err.Error())
	}

	err = ReplaceFile(s.replaceFile.Name(), s.replaceFile1.Name(), fileStat.Mode, logrus.NewEntry(logrus.StandardLogger()))
	if err != nil {
		t.Error(err.Error())
	}
}
