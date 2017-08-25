/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package log

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestSetLevel(t *testing.T) {
	text := &bytes.Buffer{}
	logrus.SetOutput(text)

	Warning("Warning message")
	Info("Info message")
	Debug("Debug message")

	lines1 := strings.Count(text.String(), "\n")

	text.Reset()
	err := SetLevel("debug")
	if err != nil {
		t.Error("Valid log level should not return an error")
	}

	Warning("Warning message")
	Info("Info message")
	Debug("Debug message")

	lines2 := strings.Count(text.String(), "\n")

	if lines1 >= lines2 {
		t.Error("Changig log level to debug failed")
	}

	err = SetLevel("abcd")
	if err == nil {
		t.Error("invalid log level should return an error")
	}
}

func TestSetOutput(t *testing.T) {
	temp, err := ioutil.TempFile("/tmp", "")
	if err != nil {
		t.Error(err)
	}

	defer temp.Close()
	defer os.Remove(temp.Name())

	if err := SetOutput(temp.Name()); err != nil {
		t.Error(err)
	}

	Warning("Warning message")
	Info("Info message")
	Debug("Debug message")

	time.Sleep(1 * time.Second)
	data := make([]byte, 100)
	read, err := temp.Read(data)
	if err != nil {
		t.Error(err)
	}
	if read <= 0 {
		t.Error("Logging to file doesn't work: file is empty after logging")
	}

	if err := SetOutput("/tmp"); err == nil {
		t.Error("Setting the log output to an existing directory should return an error")
	}
}

func TestSetFormatter(t *testing.T) {
	text := &bytes.Buffer{}
	json := &bytes.Buffer{}

	logrus.SetOutput(text)
	Info("Info message")
	Error("Error message")

	SetFormatter("json")
	logrus.SetOutput(json)
	Info("Info message")
	Error("Error message")

	if !(len(json.String()) > len(text.String())) {
		t.Error("JSON logging doesn't seem to work")
	}
	SetFormatter("text")
}
