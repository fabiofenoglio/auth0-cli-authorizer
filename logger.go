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
	Debugf(msg string, args ...interface{})
	Infof(msg string, args ...interface{})
	Warningf(msg string, args ...interface{})
	Errorf(msg string, args ...interface{})
}

type consoleLogger struct{}

var _ Logger = &consoleLogger{}

func (c consoleLogger) Debugf(msg string, args ...interface{}) {
	fmt.Println("[debug] " + fmt.Sprintf(msg, args...))
}

func (c consoleLogger) Infof(msg string, args ...interface{}) {
	fmt.Println("[info]  " + fmt.Sprintf(msg, args...))
}

func (c consoleLogger) Warningf(msg string, args ...interface{}) {
	fmt.Println("[WARN]  " + fmt.Sprintf(msg, args...))
}

func (c consoleLogger) Errorf(msg string, args ...interface{}) {
	fmt.Println("[ERROR] " + fmt.Sprintf(msg, args...))
}

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

func (c noOpLogger) Debugf(_ string, _ ...interface{}) {
	// NOP
}

func (c noOpLogger) Infof(_ string, _ ...interface{}) {
	// NOP
}

func (c noOpLogger) Warningf(_ string, _ ...interface{}) {
	// NOP
}

func (c noOpLogger) Errorf(_ string, _ ...interface{}) {
	// NOP
}

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
