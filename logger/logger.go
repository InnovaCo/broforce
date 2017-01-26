package logger

import (
	"io"
	"time"

	"github.com/Sirupsen/logrus"
)

func New(w io.Writer, level string) *logrus.Logger {
	Log = logrus.StandardLogger()
	logrus.SetOutput(w)
	Log.Formatter = &logrus.TextFormatter{TimestampFormat: time.RFC3339, FullTimestamp: true}
	if lvl, err := logrus.ParseLevel(level); err == nil {
		Log.Level = lvl
	}
	return Log
}

var Log *logrus.Logger
