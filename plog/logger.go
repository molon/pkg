package plog

import (
	"github.com/sirupsen/logrus"
)

type Logger interface {
	Debug(args ...interface{})
	Debugln(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infoln(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnln(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorln(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalln(args ...interface{})
	Fatalf(format string, args ...interface{})
}

func SetLogger(l Logger) {
	logger = l
}

func newLogger() Logger {
	l := logrus.StandardLogger()
	l.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000000",
	})
	return l
}
