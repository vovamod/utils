package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
)

var (
	std      *AllLog = New(os.Stdout, LoggerInfo)
	globalMu sync.RWMutex
)

func Default() *AllLog {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return std
}
func Replace(l *AllLog) {
	globalMu.Lock()
	std = l
	globalMu.Unlock()
}
func SetOutput(out io.Writer) { std.SetOutput(out) }
func SetType(t LoggerType)    { std.SetType(t) }
func SetFlags(f int)          { std.SetFlags(f) }
func SetDepth(d int)          { std.SetDepth(d) }

func Debug(v ...any)   { std.Debug(v...) }
func Info(v ...any)    { std.Info(v...) }
func Warn(v ...any)    { std.Warn(v...) }
func Error(v ...any)   { std.Error(v...) }
func Fatal(v ...any)   { std.Fatal(v...) }
func Success(v ...any) { std.Success(v...) }
func Notice(v ...any)  { std.Notice(v...) }

func Debugf(f string, v ...any)                          { std.Debugf(f, v...) }
func Infof(f string, v ...any)                           { std.Infof(f, v...) }
func Warnf(f string, v ...any)                           { std.Warnf(f, v...) }
func Errorf(f string, v ...any)                          { std.Errorf(f, v...) }
func Fatalf(f string, v ...any)                          { std.Fatalf(f, v...) }
func Successf(f string, v ...any)                        { std.Successf(f, v...) }
func Noticef(f string, v ...any)                         { std.Noticef(f, v...) }
func Customf(levelName string, f string, v ...any)       { std.Customf(levelName, f, v...) }
func Streamf(f string, v ...any)                         { std.Streamf(f, v...) }
func CustomStreamf(levelName string, f string, v ...any) { std.CustomStreamf(levelName, f, v...) }

func WithField(k string, v any) *Entry { return std.WithField(k, v) }
func WithFields(f Fields) *Entry       { return std.WithFields(f) }

// New instance of AllLog
func New(out io.Writer, tp LoggerType) *AllLog {
	return &AllLog{
		slog:   log.New(out, "", log.Ldate|log.Lmicroseconds),
		tp:     tp,
		depth:  0,
		exitFn: os.Exit,
	}
}

// SetOutput - Set output for logs
func (l *AllLog) SetOutput(output io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.slog.SetOutput(output)
}

// SetDepth - Set depth to look for file. If 0 no filename will be listed in log
func (l *AllLog) SetDepth(depth int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.slog.Flags() != log.Lshortfile && l.slog.Flags() != log.Llongfile {
		fmt.Print("WARNING. YOU DO NOT HAVE A FLAG SPECIFIED IN slog INSTANCE THAT ENABLES depth SUPPORT.")
	}
	l.depth = depth
}

// SetType - Set type of log to look for
func (l *AllLog) SetType(t LoggerType) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if t.IsValid() {
		l.tp = t
		return
	}
	fmt.Printf("Logger type %v is invalid. Defaulting to INFO\n", t)
}

// SetFlags - provide log.L* flags here.
func (l *AllLog) SetFlags(value int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.slog.SetFlags(value)
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

func (l *AllLog) Debug(v ...any) {
	_ = l.createPerCall(LoggerDebug, "", v)
}

func (l *AllLog) Info(v ...any) {
	_ = l.createPerCall(LoggerInfo, "", v)
}

func (l *AllLog) Warn(v ...any) {
	_ = l.createPerCall(LoggerWarn, "", v)
}

func (l *AllLog) Error(v ...any) {
	_ = l.createPerCall(LoggerError, "", v)
}

func (l *AllLog) Fatal(v ...any) {
	_ = l.createPerCall(LoggerFatal, "", v)
	l.exitFn(1)
}

func (l *AllLog) Success(v ...any) {
	_ = l.createPerCall(LoggerSuccess, "", v)
}

func (l *AllLog) Notice(v ...any) {
	_ = l.createPerCall(LoggerNotice, "", v)
}

func (l *AllLog) Debugf(format string, v ...any) {
	_ = l.createPerCall(LoggerDebug, format, v)
}

func (l *AllLog) Infof(format string, v ...any) {
	_ = l.createPerCall(LoggerInfo, format, v)
}

func (l *AllLog) Warnf(format string, v ...any) {
	_ = l.createPerCall(LoggerWarn, format, v)
}

func (l *AllLog) Errorf(format string, v ...any) {
	_ = l.createPerCall(LoggerError, format, v)
}

func (l *AllLog) Fatalf(format string, v ...any) {
	_ = l.createPerCall(LoggerFatal, format, v)
	l.exitFn(1)
}

func (l *AllLog) Successf(format string, v ...any) {
	_ = l.createPerCall(LoggerSuccess, format, v)
}

func (l *AllLog) Noticef(format string, v ...any) {
	_ = l.createPerCall(LoggerNotice, format, v)
}

func (l *AllLog) Customf(levelName string, format string, v ...any) {
	prefixVal, ok := customLevels.Load(levelName)
	prefix := ColorCyan + "[" + strings.ToUpper(levelName) + "]" + ColorReset
	if ok {
		prefix = prefixVal.(string)
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	message := fmt.Sprintf(format, v...)
	l.isStreaming = false
	d := l.depth
	_ = l.slog.Output(d, prefix+message)
}

// Streamf - an ability to stream message
func (l *AllLog) Streamf(format string, v ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.tp > LoggerInfo {
		return
	}

	message := fmt.Sprintf(format, v...)
	prefix := ColorBrightGreen + "[STREAM]" + ColorReset

	if l.isStreaming {
		fmt.Print("\033[1A\033[2K") // mv1up & clear full
	}

	_ = l.slog.Output(l.depth, prefix+message)

	l.isStreaming = true
}

// CustomStreamf - add a custom log level for streaming
func (l *AllLog) CustomStreamf(levelName string, format string, v ...any) {
	prefixVal, ok := customLevels.Load(levelName)
	prefix := ColorCyan + "[" + strings.ToUpper(levelName) + "]" + ColorReset
	if ok {
		prefix = prefixVal.(string)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.tp > LoggerInfo {
		return
	}
	message := fmt.Sprintf(format, v...)

	if l.isStreaming {
		fmt.Print("\033[1A\033[2K")
	}

	_ = l.slog.Output(l.depth, prefix+message)
	l.isStreaming = true
}

// General

func (l *AllLog) createPerCall(tp LoggerType, format string, v []any) string {
	l.mu.Lock()
	if tp < l.tp {
		l.mu.Unlock()
		return ""
	}
	l.isStreaming = false
	l.mu.Unlock()

	var message string
	if len(v) > 0 && len(format) > 0 {
		message = fmt.Sprintf(format, v...)
	} else {
		message = fmt.Sprint(v...)
	}

	finalMsg := tp.toString() + message

	_ = l.slog.Output(l.depth, finalMsg)

	if l.flog != nil {
		_ = l.flog.Output(l.depth, finalMsg)
	}

	return finalMsg
}

// TESTING
type Fields map[string]any
type Entry struct {
	logger *AllLog
	fields Fields
}

func (l *AllLog) WithField(key string, value any) *Entry {
	return &Entry{
		logger: l,
		fields: Fields{key: value},
	}
}

func (l *AllLog) WithFields(f Fields) *Entry {
	return &Entry{
		logger: l,
		fields: f,
	}
}

func (e *Entry) formatFields() string {
	if len(e.fields) == 0 {
		return ""
	}

	keys := make([]string, 0, len(e.fields))
	for k := range e.fields {
		keys = append(keys, k)
	}
	sort.Strings(keys) // Keep logs consistent

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s%s%s=%v", ColorCyan, k, ColorReset, e.fields[k]))
	}
	return " " + strings.Join(parts, " ")
}

func (e *Entry) log(tp LoggerType, format string, v ...any) {
	if e.logger.tp > tp {
		return
	}

	var msg string
	if format == "" {
		msg = fmt.Sprint(v...)
	} else {
		msg = fmt.Sprintf(format, v...)
	}

	msg += e.formatFields()

	_ = e.logger.slog.Output(e.logger.depth, tp.toString()+msg)
}

func (e *Entry) WithField(key string, value any) *Entry {
	return e.WithFields(Fields{key: value})
}

func (e *Entry) WithFields(f Fields) *Entry {
	newFields := make(Fields, len(e.fields)+len(f))

	for k, v := range e.fields {
		newFields[k] = v
	}
	for k, v := range f {
		newFields[k] = v
	}

	return &Entry{
		logger: e.logger,
		fields: newFields,
	}
}

func (e *Entry) Info(v ...any)                    { e.log(LoggerInfo, "", v...) }
func (e *Entry) Error(v ...any)                   { e.log(LoggerError, "", v...) }
func (e *Entry) Debug(v ...any)                   { e.log(LoggerDebug, "", v...) }
func (e *Entry) Warn(v ...any)                    { e.log(LoggerWarn, "", v...) }
func (e *Entry) Success(v ...any)                 { e.log(LoggerSuccess, "", v...) }
func (e *Entry) Infof(format string, v ...any)    { e.log(LoggerInfo, format, v...) }
func (e *Entry) Errorf(format string, v ...any)   { e.log(LoggerError, format, v...) }
func (e *Entry) Debugf(format string, v ...any)   { e.log(LoggerDebug, format, v...) }
func (e *Entry) Warnf(format string, v ...any)    { e.log(LoggerWarn, format, v...) }
func (e *Entry) Successf(format string, v ...any) { e.log(LoggerSuccess, format, v...) }
