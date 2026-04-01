package log

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
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

	Streamf(format string, v ...any)
}

// Setup of logger

// SetOutput - Set output for logs
func SetOutput(output io.Writer) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.slog.SetOutput(output)
}

// SetDepth - Set depth to look for file. If 0 no filename will be listed in log
func SetDepth(depth int) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.depth = depth
}

// SetType - Set type of log to look for
func SetType(t LoggerType) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	if t.IsValid() {
		logger.tp = t
		return
	}
	fmt.Printf("Logger type %v is invalid. Defaulting to INFO\n", t)
}

// SetFlags - provide log.L* flags here.
func SetFlags(value int) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.slog.SetFlags(value)
}

// RegisterCustom - register your log level. You can specify format: [MESSAGE] where MESSAGE must be %s so that the name of your custom level would be there
func RegisterCustom(name string, colorCode string, format *string) {
	var cm string
	if format != nil {
		cm = fmt.Sprintf(*format, name)
	} else {
		cm = colorCode + "[" + strings.ToUpper(name) + "]" + ColorReset
	}
	customLevels.Store(name, cm)
}

// Levels of logging

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

func Customf(levelName string, format string, v ...any) {
	prefixVal, ok := customLevels.Load(levelName)
	prefix := ColorCyan + "[" + strings.ToUpper(levelName) + "]" + ColorReset
	if ok {
		prefix = prefixVal.(string)
	}

	logger.mu.Lock()
	defer logger.mu.Unlock()
	message := fmt.Sprintf(format, v...)
	logger.isStreaming = false
	d := logger.depth
	_ = logger.slog.Output(d, prefix+message)
}

// Streamf - an ability to stream message
func Streamf(format string, v ...any) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	if logger.tp > LoggerInfo {
		return
	}

	message := fmt.Sprintf(format, v...)
	prefix := ColorBrightGreen + "[STREAM]" + ColorReset

	if logger.isStreaming {
		fmt.Print("\033[1A\033[2K") // mv1up & clear full
	}

	_ = logger.slog.Output(logger.depth, prefix+message)

	logger.isStreaming = true
}

// CustomStreamf - add a custom log level for streaming
func CustomStreamf(levelName string, format string, v ...any) {
	prefixVal, ok := customLevels.Load(levelName)
	prefix := ColorCyan + "[" + strings.ToUpper(levelName) + "]" + ColorReset
	if ok {
		prefix = prefixVal.(string)
	}

	logger.mu.Lock()
	defer logger.mu.Unlock()

	if logger.tp > LoggerInfo {
		return
	}
	message := fmt.Sprintf(format, v...)

	if logger.isStreaming {
		fmt.Print("\033[1A\033[2K")
	}

	_ = logger.slog.Output(logger.depth, prefix+message)
	logger.isStreaming = true
}

// General

func CreatePerCall(tp LoggerType, format string, v ...any) string {
	logger.mu.Lock()
	if logger.tp > tp {
		logger.mu.Unlock()
		return ""
	}

	logger.isStreaming = false
	d := logger.depth
	logger.mu.Unlock()

	var message string
	if len(v) > 0 {
		message = fmt.Sprintf(format, v...)
	} else {
		message = format
	}

	var buffer bytes.Buffer
	buffer.WriteString(tp.toString())
	_, _ = fmt.Fprint(&buffer, message)

	finalMsg := buffer.String()
	_ = logger.slog.Output(d, finalMsg)

	if logger.flog != nil {
		_ = logger.flog.Output(d, finalMsg)
	}
	return finalMsg
}
