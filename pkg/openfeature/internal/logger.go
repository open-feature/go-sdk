package internal

import (
	"log"

	"github.com/go-logr/logr"
)

// Logger is the sdk's default logr.LogSink implementation.
// Logs using this logger logs only on error, all other logs are no-ops
type Logger struct{}

func (l Logger) Init(info logr.RuntimeInfo) {}

func (l Logger) Enabled(level int) bool { return true }

func (l Logger) Info(level int, msg string, keysAndValues ...interface{}) {}

func (l Logger) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Println("openfeature:", err)
}

func (l Logger) WithValues(keysAndValues ...interface{}) logr.LogSink { return l }

func (l Logger) WithName(name string) logr.LogSink { return l }
