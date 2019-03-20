// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

package logging

import (
	"github.com/sirupsen/logrus"
)

// Use golang's standard logger by default.
var logger Logger

func init() {
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:true,
		TimestampFormat: "2006/Jan/02 03:04:05",
	})

	logger = l
}

// Logger mimics golang's standard Logger as an interface.
type Logger interface {
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	Trace(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Print(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})

	Traceln(args ...interface{})
	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Println(args ...interface{})
	Warnln(args ...interface{})
	Warningln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
}

// SetLogger sets the logger to be used
func SetLogger(l Logger) {
	logger = l
}

// CurrentLogger gets the logger to be used
func CurrentLogger() Logger {
	return logger
}

func Tracef(format string, args ...interface{}) {
	logger.Tracef(format, args...)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Warningf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit() with a non-zero exit code.
func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

func Trace(args ...interface{}) {
	logger.Trace(args...)
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Warning(args ...interface{}) {
	logger.Warn(args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

func Traceln(args ...interface{}) {
	logger.Traceln(args...)
}

func Debugln(args ...interface{}) {
	logger.Debugln(args...)
}

func Infoln(args ...interface{}) {
	logger.Infoln(args...)
}

func Warnln(args ...interface{}) {
	logger.Warnln(args...)
}

func Warningln(args ...interface{}) {
	logger.Warn(args...)
}

func Errorln(args ...interface{}) {
	logger.Errorln(args...)
}

func Fatalln(args ...interface{}) {
	logger.Fatalln(args...)
}
