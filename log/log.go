package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

// ANSI color codes + logger types
const (
	ColorReset        = "\033[0m"
	ColorRed          = "\033[31m"
	ColorGreen        = "\033[32m"
	ColorYellow       = "\033[33m"
	ColorBlue         = "\033[34m"
	ColorPurple       = "\033[35m"
	ColorCyan         = "\033[36m"
	ColorGray         = "\033[37m"
	ColorWhite        = "\033[97m"
	ColorBrightRed    = "\033[91m"
	ColorBrightGreen  = "\033[92m"
	ColorBrightYellow = "\033[93m"
	ColorBrightBlue   = "\033[94m"
	ColorBrightPurple = "\033[95m"
	ColorBrightCyan   = "\033[96m"
	ColorBgRed        = "\033[41m"
	ColorBgDarkRed    = "\033[48;5;88m"

	LoggerAll   LoggerType = "ALL"
	LoggerDebug LoggerType = "DEBUG"
	LoggerInfo  LoggerType = "INFO"
	LoggerWarn  LoggerType = "WARN"
	LoggerError LoggerType = "ERROR"
	LoggerFatal LoggerType = "FATAL"
)

// LoggerType - for validation
type LoggerType string

func (lT LoggerType) IsValid() bool {
	switch lT {
	case LoggerAll, LoggerDebug, LoggerInfo, LoggerWarn, LoggerError, LoggerFatal:
		return true
	default:
		return false
	}
}

// logger - default logger instance
var logger AllLog = AllLog{
	Output: os.Stdout,
	Type:   LoggerInfo,
	Depth:  0, // if 0 - shows nothing, only shows on debug
}

// AllLog - represents logger and it's interface with params
type AllLog struct {
	Output io.Writer
	Depth  int
	Type   LoggerType
	Logger
}

type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	// Non-standard (for my bill-go-rc2)
	Success(format string, args ...interface{})
	Notice(format string, args ...interface{})
}

// SetOutput - Set output for logs
func (l *AllLog) SetOutput(output io.Writer) {
	logger.Output = output
}

// SetDepth - Set depth to look for file. If 0 no filename will be listed in log
func (l *AllLog) SetDepth(depth int) {
	logger.Depth = depth
}

func (l *AllLog) SetType(t LoggerType) {
	if t.IsValid() {
		logger.Type = t
		return
	}
	Warn("Logger type %s is invalid. Defaulting to INFO", t)
}

// getCallerInfo returns the file and line number of the caller
func getCallerInfo(depth int) string {
	if depth == 0 {
		return ""
	}

	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		return "???:0"
	}

	// Extract just the filename from the full path
	parts := strings.Split(file, "/")
	if len(parts) > 0 {
		file = parts[len(parts)-1]
	}

	return fmt.Sprintf("%s:%d", file, line)
}

// getTimestamp returns current timestamp in a nice format
func getTimestamp() string {
	return time.Now().Format("15:04:05.000")
}

// isTerminal checks if the output is a terminal
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		fileInfo, _ := f.Stat()
		return (fileInfo.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

func Debug(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(logger.Output, CreatePerCall("DEBUG", format, args...))
}

func Info(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(logger.Output, CreatePerCall("INFO", format, args...))
}

func Warn(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(logger.Output, CreatePerCall("WARN", format, args...))
}

func Error(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(logger.Output, CreatePerCall("ERROR", format, args...))
}

func Fatal(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(logger.Output, CreatePerCall("FATAL", format, args...))

	panic("Called panic due to fatal error")
}

func Success(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(logger.Output, CreatePerCall("SUCCESS", format, args...))
}

func Notice(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(logger.Output, CreatePerCall("NOTICE", format, args...))
}

// General

func CreatePerCall(call, format string, args ...interface{}) string {
	timestamp := getTimestamp()
	caller := getCallerInfo(logger.Depth)
	var message string
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	} else {
		message = format
	}

	var perCallFormat string
	switch call {
	case "DEBUG":
		perCallFormat = fmt.Sprintf("%s[%s]%s %s%s%s ", ColorGray, timestamp, ColorReset,
			ColorBlue, "DEBUG", ColorReset)
	case "INFO":
		perCallFormat = fmt.Sprintf("%s[%s]%s %s%s%s ", ColorGray, timestamp, ColorReset,
			ColorBrightGreen, "INFO", ColorReset)
	case "WARN":
		perCallFormat = fmt.Sprintf("%s[%s]%s %s%s%s ", ColorGray, timestamp, ColorReset,
			ColorBrightYellow, "WARN", ColorReset)
	case "ERROR":
		perCallFormat = fmt.Sprintf("%s[%s]%s %s%s%s ", ColorGray, timestamp, ColorReset,
			ColorBrightRed, "ERROR", ColorReset)
	case "FATAL":
		perCallFormat = fmt.Sprintf("%s[%s]%s %s%s%s ", ColorGray, timestamp, ColorReset,
			ColorBgDarkRed, "FATAL", ColorReset)
	case "SUCCESS":
		perCallFormat = fmt.Sprintf("%s[%s]%s %s%s%s ", ColorGray, timestamp, ColorReset,
			ColorGreen, "SUCCESS", ColorReset)
	case "NOTICE":
		perCallFormat = fmt.Sprintf("%s[%s]%s %s%s%s ", ColorGray, timestamp, ColorReset,
			ColorYellow, "NOTICE", ColorReset)
	}

	if isTerminal(logger.Output) {
		if call == "DEBUG" {
			return fmt.Sprintf(perCallFormat+"%s | %s", message, caller)
		}
		return fmt.Sprintf(perCallFormat+"%s\n", message)
	}
	if call == "DEBUG" {
		return fmt.Sprintf("[%s]: %s: %s | %s\n",
			timestamp, "DEBUG", message, caller)
	}
	return fmt.Sprintf("[%s]: %s: %s\n",
		timestamp, call, message)
}
