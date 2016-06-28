package fileutil

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/cloudflare/cfssl/log"
)

func init() {
	log.Level = log.LevelWarning
}

func testSameFile(t *testing.T, srcTxt, dstTxt string) bool {
	src, err := ioutil.TempFile("", "src")
	if err != nil {
		t.Errorf(err.Error())
	}
	defer os.Remove(src.Name())
	dst, err := ioutil.TempFile("", "dest")
	if err != nil {
		t.Errorf(err.Error())
	}
	defer os.Remove(dst.Name())

	_, err = src.WriteString(srcTxt)
	if err != nil {
		t.Errorf(err.Error())
	}
	_, err = dst.WriteString(dstTxt)
	if err != nil {
		t.Errorf(err.Error())
	}

	status, err := SameFile(src.Name(), dst.Name())
	if err != nil {
		t.Errorf(err.Error())
	}
	return status
}

func TestSameFileTrue(t *testing.T) {
	status := testSameFile(t, "bla", "bla")
	if status != true {
		t.Errorf("Expected SameFile(src, dest) to be %v, got %v", true, status)
	}
}

func TestSameFileFalse(t *testing.T) {
	status := testSameFile(t, "bla", "bla2")
	if status != false {
		t.Errorf("Expected sameConfig(src, dest) to be %v, got %v", false, status)
	}
}
