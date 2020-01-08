package logrusr

import (
	"fmt"
	"github.com/activeshadow/logr/util"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	verbosity int
	loggers   []string
)

func SetVerbosity(v int) {
	verbosity = v
}

func LimitToLoggers(names ...string) {
	loggers = append(loggers, names...)
}

type clock struct {
	mock time.Time // set in tests
}

func (this *clock) now() time.Time {
	if this == nil { // normal operations
		return time.Now()
	}

	return this.mock // testing only
}

type LogrusInfoLogr struct {
	enabled bool
	name    string
	kvs     map[string]interface{}
	logger  logrus.Logger

	// Normally nil, set by test code only.
	clock *clock
}

func (this LogrusInfoLogr) Info(msg string, kvs ...interface{}) {
	if !this.enabled {
		return
	}

	logger := this.logger.WithFields(logrus.Fields{
		"request": &logrus.Fields{
			"name": this.name,
			"kvs":  kvs,
			// "this_kvs": this.kvs,
		},
	})

	if logger != nil {
		logger.Info(msg)
	}

}

func (this LogrusInfoLogr) Enabled() bool {
	return this.enabled
}

type LogrusLogr struct {
	LogrusInfoLogr
}

func (this LogrusLogr) Error(err error, msg string, kvs ...interface{}) {

	logger := this.logger.WithFields(logrus.Fields{
		"request": &logrus.Fields{
			"error": err,
			"name":  this.name,
			"kvs":   kvs,
			// "this_kvs": this.kvs,
		},
	})

	if logger != nil {
		logger.Error(msg)
	}

}

func (this LogrusLogr) V(level int) logr.InfoLogger {
	if level <= verbosity {
		if len(loggers) == 0 || util.StringSliceContains(loggers, this.name) {
			return this.WithValues("v", level)
		}
	}

	return &LogrusInfoLogr{enabled: false}
}

func (this LogrusLogr) WithValues(kvs ...interface{}) logr.Logger {
	newKVs := make(map[string]interface{})

	for k, v := range this.kvs {
		newKVs[k] = v
	}

	for i := 0; i < len(kvs); i += 2 {
		newKVs[kvs[i].(string)] = kvs[i+1]
	}

	return &LogrusLogr{
		LogrusInfoLogr: LogrusInfoLogr{
			enabled: this.enabled,
			name:    this.name,
			kvs:     newKVs,
			logger:  this.logger,
		},
	}
}

func (this LogrusLogr) WithName(name string) logr.Logger {
	name = fmt.Sprintf("%s.%s", this.name, name)

	return &LogrusLogr{
		LogrusInfoLogr: LogrusInfoLogr{
			enabled: this.enabled,
			name:    name,
			kvs:     this.kvs,
			logger:  this.logger,
		},
	}
}

func New(name string, logger logrus.Logger) logr.Logger {
	return &LogrusLogr{
		LogrusInfoLogr: LogrusInfoLogr{
			enabled: true,
			name:    name,
			kvs:     make(map[string]interface{}),
			logger:  logger,
		},
	}
}
