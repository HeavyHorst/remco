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
	"github.com/hashicorp/go-hclog"
	"os"
	"sync"
)

var logger hclog.Logger
var lock sync.RWMutex

func init() {
	InitializeLogging("text", "info")
}

func InitializeLogging(format string, level string) {
	lock.Lock()
	defer lock.Unlock()

	logger = hclog.
		New(&hclog.LoggerOptions{JSONFormat: format == "json", Level: hclog.LevelFromString(level)}).
		With("prefix", fmt.Sprintf("%s[%d]", os.Args[0], os.Getpid()))
}

// Debug logs a message with severity DEBUG.
func Debug(msg string, v ...interface{}) {
	logger.Debug(msg, v...)
}

// Error logs a message with severity ERROR.
func Error(msg string, v ...interface{}) {
	logger.Error(msg, v...)
}

// Fatal logs a message with severity ERROR followed by a call to os.Exit().
func Fatal(msg string, v ...interface{}) {
	logger.Error(msg, v...)
	os.Exit(1)
}

// Info logs a message with severity INFO.
func Info(msg string, v ...interface{}) {
	logger.Info(msg, v...)
}

// Warning logs a message with severity WARNING.
func Warning(msg string, v ...interface{}) {
	logger.Warn(msg, v...)
}

func WithFields(v ...interface{}) hclog.Logger {
	return logger.With(v...)
}
