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
	log.SetFormatter(new(prefixed.TextFormatter))
	log.SetLevel(log.DebugLevel)
	logger = log.WithFields(log.Fields{
		"prefix":   "remco",
		"hostname": host,
	})
}

// SetLevel sets the log level. Valid levels are panic, fatal, error, warn, info and debug.
func SetLevel(level string) error {
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
	defer lock.Unlock()
	logger.Debug(v)
}

// Error logs a message with severity ERROR.
func Error(v ...interface{}) {
	lock.Lock()
	defer lock.Unlock()
	logger.Error(v)
}

// Fatal logs a message with severity ERROR followed by a call to os.Exit().
func Fatal(v ...interface{}) {
	lock.Lock()
	defer lock.Unlock()
	logger.Fatal(v)
}

// Info logs a message with severity INFO.
func Info(v ...interface{}) {
	lock.Lock()
	defer lock.Unlock()
	logger.Info(v)
}

// Warning logs a message with severity WARNING.
func Warning(v ...interface{}) {
	lock.Lock()
	defer lock.Unlock()
	logger.Warning(v)
}
