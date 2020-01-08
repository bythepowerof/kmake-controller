package logrusr

import (
	"errors"
	"github.com/sirupsen/logrus"
	"testing"
	"time"
)

func TestInfoLogger(t *testing.T) {
	mock, _ := time.Parse("2006-01-02", "2015-12-15")
	l := logrus.New()

	logger := New("foo", *l)
	logger.(*LogrusLogr).clock = &clock{mock: mock}
	logger.Info("test log", "hello", "world")
}

func TestErrLogger(t *testing.T) {
	mock, _ := time.Parse("2006-01-02", "2015-12-15")
	err := errors.New("BOOM SUCKA!")
	l := logrus.New()

	logger := New("bar", *l)
	logger.(*LogrusLogr).clock = &clock{mock: mock}
	logger.Error(err, "test error log", "hello", "world")
}

func TestNonVerboseLogger(t *testing.T) {
	mock, _ := time.Parse("2006-01-02", "2015-12-15")
	l := logrus.New()

	logger := New("sucka", *l)
	logger.(*LogrusLogr).clock = &clock{mock: mock}
	logger.V(1).Info("test verbose log", "hello", "crazy world")
}

func TestVerboseLogger(t *testing.T) {
	mock, _ := time.Parse("2006-01-02", "2015-12-15")

	SetVerbosity(1)

	l := logrus.New()

	logger := New("sucka", *l)

	vLogger := logger.V(1)
	vLogger.(*LogrusLogr).clock = &clock{mock: mock}
	vLogger.Info("test verbose log", "hello", "crazy world")
}

func TestNamedLogger(t *testing.T) {
	mock, _ := time.Parse("2006-01-02", "2015-12-15")

	l := logrus.New()

	logger := New("foo", *l)
	namedLogger := logger.WithName("bar")
	namedLogger.(*LogrusLogr).clock = &clock{mock: mock}
	namedLogger.Info("test log", "hello", "world")
}

func TestValuesLogger(t *testing.T) {
	mock, _ := time.Parse("2006-01-02", "2015-12-15")

	l := logrus.New()

	logger := New("foo", *l)
	valuesLogger := logger.WithValues("goodbye", "crazy world")
	valuesLogger.(*LogrusLogr).clock = &clock{mock: mock}
	valuesLogger.Info("test log", "hello", "world")
}

func TestLimitedLogger(t *testing.T) {
	mock, _ := time.Parse("2006-01-02", "2015-12-15")

	SetVerbosity(1)
	LimitToLoggers("bar")

	l := logrus.New()

	logger := New("foo", *l)
	vLogger := logger.V(1)
	vLogger.(*LogrusInfoLogr).clock = &clock{mock: mock}
	vLogger.Info("test verbose log", "hello", "crazy world")
}
