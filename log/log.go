package log

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

// logger - default logger instance
var logger = AllLog{
	slog:  log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds),
	tp:    LoggerInfo,
	depth: 0, // if 0 - shows nothing, only shows on debug
}

type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Success(format string, args ...interface{})
	Notice(format string, args ...interface{})
}

// SetOutput - Set output for logs
func (l *AllLog) SetOutput(output io.Writer) {
	logger.slog.SetOutput(output)
}

// SetDepth - Set depth to look for file. If 0 no filename will be listed in log
func (l *AllLog) SetDepth(depth int) {
	logger.depth = depth
}

// SetType - Set type of log to look for
func (l *AllLog) SetType(t LoggerType) {
	if t.IsValid() {
		logger.tp = t
		return
	}
	Warn("Logger type %v is invalid. Defaulting to INFO", t)
}

// SetFlags - provide log.L* flags here.
func (l *AllLog) SetFlags(value int) {
	logger.slog.SetFlags(value)
}

func Debug(format string, args ...interface{}) {
	_ = CreatePerCall(LoggerDebug, format, args...)
}

func Info(format string, args ...interface{}) {
	_ = CreatePerCall(LoggerInfo, format, args...)
}

func Warn(format string, args ...interface{}) {
	_ = CreatePerCall(LoggerWarn, format, args...)
}

func Error(format string, args ...interface{}) {
	_ = CreatePerCall(LoggerError, format, args...)
}

func Fatal(format string, args ...interface{}) {
	result := CreatePerCall(LoggerFatal, format, args...)

	panic(result)
}

func Success(format string, args ...interface{}) {
	_ = CreatePerCall(LoggerSuccess, format, args...)
}

func Notice(format string, args ...interface{}) {
	_ = CreatePerCall(LoggerNotice, format, args...)
}

// General

func CreatePerCall(tp LoggerType, format string, args ...interface{}) string {
	if logger.tp > tp {
		return ""
	}
	var message string
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	} else {
		message = format
	}

	var buffer bytes.Buffer
	defer buffer.Reset() // reset after each use
	buffer.WriteString(tp.toString())
	_, _ = fmt.Fprint(&buffer, message) // err can be ignored!!!

	// output result
	_ = logger.slog.Output(logger.depth, buffer.String())
	return buffer.String()
}
