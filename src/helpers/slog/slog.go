package slog

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"
)

type Logger struct {
	Name string
}

var DebugEnabled = os.Getenv("DEBUG_LOGGING") == "true"

func NewLogger(name string) Logger {
	return Logger{Name: name}
}

const (
	colorRed    = "\033[0;31m"
	colorYellow = "\033[0;33m"
	colorNone   = "\033[0m"
)

// Type - date - message - errorMsg
const errLogFormat = "%s%s - %s - %s - '%s' - '%s'%s\n"

func (l *Logger) Error(err error, msg string, a ...any) {
	message := fmt.Sprintf(msg, a...)
	_, file, line, _ := runtime.Caller(2)
	shortFile := file

	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			shortFile = file[i+1:]
			break
		}
	}

	logMessage := fmt.Sprintf("%s:%d %s", shortFile, line, message)

	fmt.Printf(errLogFormat, colorRed, l.Name, time.Now().Format("2006-01-02 15:04:05"), "ERROR", logMessage, err.Error(), colorNone)
}

// COLOR Type - date - message RESETCOLOR
const warnLogFormat = "%s %s - %s - %s \n%s %s\n"

func (l *Logger) Warn(msg string, a ...any) {
	message := fmt.Sprintf(msg, a...)
	fmt.Printf(warnLogFormat, colorYellow, time.Now().Format("2006-01-02 15:04:05"), "WARN", l.Name, message, colorNone)
}

// Type - date - message
const infoLogFormat = "%s - %s - %s - %s\n"

func (l *Logger) Info(msg string, a ...any) {
	message := fmt.Sprintf(msg, a...)
	fmt.Printf(infoLogFormat, time.Now().Format("2006-01-02 15:04:05"), "INFO", l.Name, message)
}

// COLOR Type - date - message RESETCOLOR
const debugLogFormat = "%s %s - %s - %s \n%s %s\n"

func (l *Logger) Debug(msg string, a ...any) {
	/*if !DebugEnabled {
		return
	}*/
	message := fmt.Sprintf(msg, a...)
	fmt.Printf(debugLogFormat, colorYellow, time.Now().Format("2006-01-02 15:04:05"), "DEBUG", l.Name, message, colorNone)
}

func (l *Logger) DebugStruct(msg string, a any) {
	if !DebugEnabled {
		return
	}
	stringStruct, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		l.Debug(msg)
		return
	}

	fmt.Printf(debugLogFormat, colorRed, time.Now().Format("2006-01-02 15:04:05"), "DEBUG", l.Name, string(stringStruct), colorNone)
}

// Type - date - code - message - errorMsg
const errHttpLogFormat = "%s - %s - %s - %d - '%s' - '%s'\n"

func (l *Logger) HttpError(err error, msg string, statusCode int, a ...any) {
	message := fmt.Sprintf(msg, a...)
	fmt.Printf(errHttpLogFormat, time.Now().Format("2006-01-02 15:04:05"), "ERROR", l.Name, statusCode, message, err)
}

// Type - date - code - message
const infoHttpLogFormat = "%s - %s - %s - %d - %s\n"

func (l *Logger) HttpInfo(msg string, statusCode int, a ...any) {
	message := fmt.Sprintf(msg, a...)
	fmt.Printf(infoHttpLogFormat, time.Now().Format("2006-01-02 15:04:05"), "INFO", l.Name, statusCode, message)
}
