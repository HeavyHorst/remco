/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package log

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var logger *log.Entry
var lock sync.RWMutex

func init() {
	SetFormatter("text")
	log.SetLevel(log.InfoLevel)
	logger = log.WithFields(log.Fields{
		"prefix": fmt.Sprintf("%s[%d]", os.Args[0], os.Getpid()),
	})
}

// SetFormatter sets the log formatter. Valid formatters are json and text.
func SetFormatter(format string) {
	if format != "" {
		lock.Lock()
		defer lock.Unlock()
		switch format {
		case "json":
			log.SetFormatter(&log.JSONFormatter{})
		case "text":
			log.SetFormatter(&prefixed.TextFormatter{DisableSorting: false})
		}
	}
}

// SetOutput sets the standard logger output to the given file.
func SetOutput(path string) error {
	if path != "" {
		lock.Lock()
		defer lock.Unlock()
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("could not open logfile %q", path)
		}
		log.SetOutput(f)
	}
	return nil
}

// SetLevel sets the log level. Valid levels are panic, fatal, error, warn, info and debug.
func SetLevel(level string) error {
	if level != "" {
		lock.Lock()
		defer lock.Unlock()
		lvl, err := log.ParseLevel(level)
		if err != nil {
			return err
		}
		log.SetLevel(lvl)
	}
	return nil
}

// some log middleware handlers

//withSource adds the field source to the logs
//example: source=resource.go:310
func withSource(l *log.Entry) *log.Entry {
	if log.GetLevel() == log.DebugLevel {

		lock.Lock()
		defer lock.Unlock()

		_, file, line, ok := runtime.Caller(2)
		if !ok {
			file = "<???>"
			line = 1
		} else {
			slash := strings.LastIndex(file, "/")
			file = file[slash+1:]
		}
		logger := l.WithField("source", fmt.Sprintf("%s:%d", file, line))
		return logger
	}
	return l
}

// Debug logs a message with severity DEBUG.
func Debug(v ...interface{}) {
	withSource(logger).Debug(v)
}

// Error logs a message with severity ERROR.
func Error(v ...interface{}) {
	withSource(logger).Error(v)
}

// Fatal logs a message with severity ERROR followed by a call to os.Exit().
func Fatal(v ...interface{}) {
	withSource(logger).Fatal(v)
}

// Info logs a message with severity INFO.
func Info(v ...interface{}) {
	withSource(logger).Info(v)
}

// Warning logs a message with severity WARNING.
func Warning(v ...interface{}) {
	withSource(logger).Warning(v)
}

// WithFields is a Wrapper for logrus.WithFields
func WithFields(fields log.Fields) *log.Entry {
	return withSource(logger).WithFields(fields)
}
