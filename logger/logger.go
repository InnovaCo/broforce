package logger

import (
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/evalphobia/logrus_fluent"

	"github.com/InnovaCo/broforce/config"
)

func New(cfg config.ConfigData) *logrus.Logger {
	Log = logrus.StandardLogger()
	Log.Formatter = &logrus.TextFormatter{TimestampFormat: time.RFC3339, FullTimestamp: true}

	if cfg.Exist("file") {
		f, err := os.OpenFile(cfg.GetStringOr("file.name", "broforce.log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		logrus.SetOutput(f)
		if lvl, err := logrus.ParseLevel(cfg.GetStringOr("file.level", "info")); err == nil {
			Log.Level = lvl
		}
	}

	if cfg.Exist("fluentd") {
		if hook, err := logrus_fluent.New(cfg.GetStringOr("fluentd.host", "localhost"), cfg.GetIntOr("fluentd.port", 24224)); err == nil {
			levels := []logrus.Level{}
			for _, lvl := range cfg.GetArray("fluentd.levels") {
				if l, err := logrus.ParseLevel(lvl.String()); err == nil {
					levels = append(levels, l)
				}
			}
			Log.Debugf("fluentd levels: %v", levels)

			hook.SetLevels(levels)
			logrus.AddHook(hook)
		} else {
			Log.Errorf("fluentd: %v", err)
		}
	}

	return Log
}

var Log *logrus.Logger
