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

const colorRed = "\033[0;31m"
const colorYellow = "\033[0;33m"
const colorNone = "\033[0m"

// Type - date - message - errorMsg
const errLogFormat = "%s%s - %s - %s - '%s' - '%s'%s\n"

func (l *Logger) Error(err error, msg string, a ...any) {
	message := fmt.Sprintf(msg, a...)
	fmt.Printf(errLogFormat, colorRed, "ERROR", l.Name, time.Now().Format("2006-01-02 15:04:05"), message, err.Error(), colorNone)
}

// Type - date - message
const infoLogFormat = "%s - %s - %s - %s\n"

func (l *Logger) Info(msg string, a ...any) {
	message := fmt.Sprintf(msg, a...)
	fmt.Printf(infoLogFormat, "INFO", time.Now().Format("2006-01-02 15:04:05"), l.Name, message)
}

// Type - date - message
const warnLogFormat = "%s%s - %s - %s - %s\n%s"

func (l *Logger) Warn(msg string, a ...any) {
	message := fmt.Sprintf(msg, a...)
	fmt.Printf(warnLogFormat, colorYellow, "INFO", time.Now().Format("2006-01-02 15:04:05"), l.Name, message, colorNone)
}

// COLOR Type - date - message RESETCOLOR
const debugLogFormat = "%s %s - %s - %s \n%s %s\n"

func (l *Logger) Debug(msg string, a ...any) {
	message := fmt.Sprintf(msg, a...)
	fmt.Printf(debugLogFormat, colorYellow, "DEBUG", time.Now().Format("2006-01-02 15:04:05"), l.Name, message, colorNone)
}

func (l *Logger) DebugStruct(msg string, a any) {
	stringStruct, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		l.Debug(msg)
		return
	}

	fmt.Printf(debugLogFormat, colorRed, "DEBUG", time.Now().Format("2006-01-02 15:04:05"), l.Name, string(stringStruct), colorNone)
}

// Type - date - code - message - errorMsg
const errHttpLogFormat = "%s - %s - %s - %d - '%s' - '%s'\n"

func (l *Logger) HttpError(err error, msg string, statusCode int, a ...any) {
	message := fmt.Sprintf(msg, a...)
	fmt.Printf(errHttpLogFormat, "ERROR", time.Now().Format("2006-01-02 15:04:05"), l.Name, statusCode, message, err)
}

// Type - date - code - message
const infoHttpLogFormat = "%s - %s - %s - %d - %s\n"

func (l *Logger) HttpInfo(msg string, statusCode int, a ...any) {
	message := fmt.Sprintf(msg, a...)
	fmt.Printf(infoHttpLogFormat, "INFO", time.Now().Format("2006-01-02 15:04:05"), l.Name, statusCode, message)
}
