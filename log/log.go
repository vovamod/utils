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
	Debug(format string)
	Info(format string)
	Warn(format string)
	Error(format string)
	Success(format string)
	Notice(format string)

	Debugf(format string, v ...any)
	Infof(format string, v ...any)
	Warnf(format string, v ...any)
	Errorf(format string, v ...any)
	Successf(format string, v ...any)
	Noticef(format string, v ...any)
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
	Warnf("Logger type %v is invalid. Defaulting to INFO", t)
}

// SetFlags - provide log.L* flags here.
func (l *AllLog) SetFlags(value int) {
	logger.slog.SetFlags(value)
}

func Debug(format string) {
	_ = CreatePerCall(LoggerDebug, format)
}

func Info(format string) {
	_ = CreatePerCall(LoggerInfo, format)
}

func Warn(format string) {
	_ = CreatePerCall(LoggerWarn, format)
}

func Error(format string) {
	_ = CreatePerCall(LoggerError, format)
}

func Fatal(format string) {
	result := CreatePerCall(LoggerFatal, format)
	// Fatal if required
	panic(result)
}

func Success(format string) {
	_ = CreatePerCall(LoggerSuccess, format)
}

func Notice(format string) {
	_ = CreatePerCall(LoggerNotice, format)
}

func Debugf(format string, v ...any) {
	_ = CreatePerCall(LoggerDebug, format, v...)
}

func Infof(format string, v ...any) {
	_ = CreatePerCall(LoggerInfo, format, v...)
}

func Warnf(format string, v ...any) {
	_ = CreatePerCall(LoggerWarn, format, v...)
}

func Errorf(format string, v ...any) {
	_ = CreatePerCall(LoggerError, format, v...)
}

func Fatalf(format string, v ...any) {
	result := CreatePerCall(LoggerFatal, format, v...)
	// Fatal if required
	panic(result)
}

func Successf(format string, v ...any) {
	_ = CreatePerCall(LoggerSuccess, format, v...)
}

func Noticef(format string, v ...any) {
	_ = CreatePerCall(LoggerNotice, format, v...)
}

// General

func CreatePerCall(tp LoggerType, format string, v ...any) string {
	if logger.tp > tp {
		return ""
	}
	var message string
	if len(v) > 0 {
		message = fmt.Sprintf(format, v...)
	} else {
		message = format
	}

	var buffer bytes.Buffer
	defer buffer.Reset() // reset after each use
	buffer.WriteString(tp.toString())
	_, _ = fmt.Fprint(&buffer, message) // err can be ignored!!!

	// output result
	_ = logger.slog.Output(logger.depth, buffer.String())
	if logger.flog != nil {
		_ = logger.flog.Output(logger.depth, buffer.String())
	}
	return buffer.String()
}
