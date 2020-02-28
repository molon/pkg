package plog

import (
	"os"

	"github.com/sirupsen/logrus"
)

func newLogger() Logger {
	l := logrus.StandardLogger()
	l.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000000",
	})
	return l
}

var logger = newLogger()

func SetLogger(l Logger) {
	logger = l
}

func GetLogger() Logger {
	return logger
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Debugln(args ...interface{}) {
	logger.Debugln(args...)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Infoln(args ...interface{}) {
	logger.Infoln(args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Warningf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Warnln(args ...interface{}) {
	logger.Warnln(args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Errorln(args ...interface{}) {
	logger.Errorln(args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
	os.Exit(1)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
	os.Exit(1)
}

func Fatalln(args ...interface{}) {
	logger.Fatalln(args...)
	os.Exit(1)
}
