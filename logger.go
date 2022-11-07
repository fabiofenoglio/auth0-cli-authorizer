package auth0cliauthorizer

import (
	"fmt"
	"strings"
)

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
}

type consoleLogger struct{}

var _ Logger = &consoleLogger{}

func (c consoleLogger) Debug(args ...interface{}) {
	fmt.Println("[debug] " + c.message(args))
}

func (c consoleLogger) Info(args ...interface{}) {
	fmt.Println("[info]  " + c.message(args))
}

func (c consoleLogger) Warning(args ...interface{}) {
	fmt.Println("[WARN]  " + c.message(args))
}

func (c consoleLogger) Error(args ...interface{}) {
	fmt.Println("[ERROR] " + c.message(args))
}

func (c consoleLogger) message(args ...interface{}) string {
	msg := ""
	for _, a := range args {
		msg += fmt.Sprintf("%v ", a)
	}
	return strings.TrimSuffix(msg, " ")
}

type noOpLogger struct{}

var _ Logger = &noOpLogger{}

func (c noOpLogger) Debug(_ ...interface{}) {
	// NOP
}

func (c noOpLogger) Info(_ ...interface{}) {
	// NOP
}

func (c noOpLogger) Warning(_ ...interface{}) {
	// NOP
}

func (c noOpLogger) Error(_ ...interface{}) {
	// NOP
}

type loggerWrapper struct {
	underlying Logger
}

func (a *loggerWrapper) Debug(args ...interface{}) {
	a.underlying.Debug(args...)
}

func (a *loggerWrapper) Info(args ...interface{}) {
	a.underlying.Info(args...)
}

func (a *loggerWrapper) Warning(args ...interface{}) {
	a.underlying.Warning(args...)
}

func (a *loggerWrapper) Error(args ...interface{}) {
	a.underlying.Error(args...)
}

func (a *loggerWrapper) Debugf(msg string, args ...interface{}) {
	a.underlying.Debug(fmt.Sprintf(msg, args...))
}

func (a *loggerWrapper) Infof(msg string, args ...interface{}) {
	a.underlying.Info(fmt.Sprintf(msg, args...))
}

func (a *loggerWrapper) Warningf(msg string, args ...interface{}) {
	a.underlying.Warning(fmt.Sprintf(msg, args...))
}

func (a *loggerWrapper) Errorf(msg string, args ...interface{}) {
	a.underlying.Error(fmt.Sprintf(msg, args...))
}
