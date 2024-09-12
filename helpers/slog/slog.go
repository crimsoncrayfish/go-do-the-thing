package slog

import (
	"encoding/json"
	"fmt"
	"time"
)

type Logger struct {
	Name string
}

func NewLogger(name string) *Logger {
	return &Logger{Name: name}

}

// Type - date - message - errorMsg
const errLogFormat = "%s - %s - %s - '%s' - '%s'\n"

func (l *Logger) Error(err error, msg string, a ...any) {
	message := fmt.Sprintf(msg, a...)
	fmt.Printf(errLogFormat, l.Name, "ERROR", time.Now().Format("2006-01-02 15:04:05"), message, err.Error())
}

// Type - date - message
const infoLogFormat = "%s - %s - %s - %s\n"

func (l *Logger) Info(msg string, a ...any) {
	message := fmt.Sprintf(msg, a...)
	fmt.Printf(infoLogFormat, l.Name, "INFO", time.Now().Format("2006-01-02 15:04:05"), message)
}

// Type - date - message
const debugLogFormat = "%s - %s - %s \n%s\n"

func (l *Logger) Debug(msg string, a ...any) {
	message := fmt.Sprintf(msg, a...)
	fmt.Printf(debugLogFormat, l.Name, "DEBUG", time.Now().Format("2006-01-02 15:04:05"), message)
}

func (l *Logger) DebugStruct(msg string, a any) {
	stringStruct, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		l.Debug(msg)
		return
	}

	fmt.Printf(debugLogFormat, l.Name, "DEBUG", time.Now().Format("2006-01-02 15:04:05"), string(stringStruct))
}

// Type - date - code - message - errorMsg
const errHttpLogFormat = "%s - %s - %s - %d - '%s' - '%s'\n"

func (l *Logger) HttpError(err error, msg string, statusCode int, a ...any) {
	message := fmt.Sprintf(msg, a...)
	fmt.Printf(errHttpLogFormat, l.Name, "ERROR", time.Now().Format("2006-01-02 15:04:05"), statusCode, message, err)
}

// Type - date - code - message
const infoHttpLogFormat = "%s - %s - %s - %d - %s\n"

func (l *Logger) HttpInfo(msg string, statusCode int, a ...any) {
	message := fmt.Sprintf(msg, a...)
	fmt.Printf(infoHttpLogFormat, l.Name, "INFO", time.Now().Format("2006-01-02 15:04:05"), statusCode, message)
}
