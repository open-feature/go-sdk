package internal

import (
	"log"

	"github.com/go-logr/logr"
)

const (
	Info  = 0
	Debug = 1
)

// Logger is the sdk's default Logger. Logs using the standard log package on error, all other logs are no-ops
type Logger struct{}

func (l Logger) Init(info logr.RuntimeInfo) {}

func (l Logger) Enabled(level int) bool { return true }

func (l Logger) Info(level int, msg string, keysAndValues ...interface{}) {}

func (l Logger) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Println("openfeature:", err)
}

func (l Logger) WithValues(keysAndValues ...interface{}) logr.LogSink { return l }

func (l Logger) WithName(name string) logr.LogSink { return l }
