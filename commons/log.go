package commons

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}

// Log logs with trace ID from context
func Log(ctx context.Context, format string, args ...interface{}) {
	tid := Tid(ctx)
	msg := fmt.Sprintf(format, args...)
	log.WithField("trace_id", tid).Info(msg)
}

// LogW logs warning with trace ID
func LogW(ctx context.Context, format string, args ...interface{}) {
	tid := Tid(ctx)
	msg := fmt.Sprintf(format, args...)
	log.WithField("trace_id", tid).Warn(msg)
}

// LogE logs error with trace ID
func LogE(ctx context.Context, format string, args ...interface{}) {
	tid := Tid(ctx)
	msg := fmt.Sprintf(format, args...)
	log.WithField("trace_id", tid).Error(msg)
}

// LogC logs critical error with trace ID
func LogC(ctx context.Context, format string, args ...interface{}) {
	tid := Tid(ctx)
	msg := fmt.Sprintf(format, args...)
	log.WithField("trace_id", tid).WithField("level", "CRITICAL").Error(msg)
}

// SetLogLevel sets the logging level
func SetLogLevel(level string) {
	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
}
