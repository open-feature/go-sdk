package openfeature

import (
	"log"

	"github.com/go-logr/logr"
)

const (
	info  = 0
	debug = 1
)

// logger is the sdk's default logger
// logs using the standard log package on error, all other logs are no-ops
type logger struct{}

func (l logger) Init(info logr.RuntimeInfo) {}

func (l logger) Enabled(level int) bool { return true }

func (l logger) Info(level int, msg string, keysAndValues ...interface{}) {}

func (l logger) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Println("openfeature:", err)
}

func (l logger) WithValues(keysAndValues ...interface{}) logr.LogSink { return l }

func (l logger) WithName(name string) logr.LogSink { return l }
