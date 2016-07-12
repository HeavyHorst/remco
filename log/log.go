package log

import (
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var logger *log.Entry
var lock sync.RWMutex

func init() {
	host, _ := os.Hostname()
	SetFormatter("text")
	log.SetLevel(log.DebugLevel)
	logger = log.WithFields(log.Fields{
		"prefix":   "remco",
		"hostname": host,
	})
}

func SetFormatter(format string) {
	lock.Lock()
	defer lock.Unlock()
	switch format {
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	case "text":
		log.SetFormatter(&prefixed.TextFormatter{DisableSorting: true})
	}
}

// SetLevel sets the log level. Valid levels are panic, fatal, error, warn, info and debug.
func SetLevel(level string) error {
	lock.Lock()
	defer lock.Unlock()
	lvl, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(lvl)
	return nil
}

// Debug logs a message with severity DEBUG.
func Debug(v ...interface{}) {
	lock.Lock()
	logger.Debug(v)
	lock.Unlock()
}

// Error logs a message with severity ERROR.
func Error(v ...interface{}) {
	lock.Lock()
	logger.Error(v)
	lock.Unlock()
}

// Fatal logs a message with severity ERROR followed by a call to os.Exit().
func Fatal(v ...interface{}) {
	lock.Lock()
	logger.Fatal(v)
	lock.Unlock()
}

// Info logs a message with severity INFO.
func Info(v ...interface{}) {
	lock.Lock()
	logger.Info(v)
	lock.Unlock()
}

// Warning logs a message with severity WARNING.
func Warning(v ...interface{}) {
	lock.Lock()
	logger.Warning(v)
	lock.Unlock()
}

// WithFields is a Wrapper for logrus.WithFields
func WithFields(fields log.Fields) *log.Entry {
	return logger.WithFields(fields)
}
