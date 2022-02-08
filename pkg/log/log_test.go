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
	"github.com/hashicorp/go-hclog"
	"strings"
	"testing"
)

func TestSetLevel(t *testing.T) {
	text := &bytes.Buffer{}
	hclog.DefaultOutput = text

	InitializeLogging("text", "info")
	Warning("Warning message")
	Info("Info message")
	Debug("Debug message")

	lines1 := strings.Count(text.String(), "\n")

	text.Reset()
	InitializeLogging("text", "debug")

	Warning("Warning message")
	Info("Info message")
	Debug("Debug message")

	lines2 := strings.Count(text.String(), "\n")

	if lines1 >= lines2 {
		t.Error("Changing log level to debug failed")
	}
}

func TestSetFormatter(t *testing.T) {
	text := &bytes.Buffer{}
	json := &bytes.Buffer{}

	hclog.DefaultOutput = text
	InitializeLogging("text", "info")
	Info("Info message")
	Error("Error message")

	hclog.DefaultOutput = json
	InitializeLogging("json", "info")

	Info("Info message")
	Error("Error message")

	txtString := text.String()
	jsonString := json.String()
	if !(len(jsonString) > len(txtString)) {
		t.Errorf("JSON logging doesn't seem to work %s vs %s", jsonString, txtString)
	}
	//fmt.Printf("$$$$$$$$$%s vs %s \n", txtString, jsonString)
	InitializeLogging("text", "info")
}
