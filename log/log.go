package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	std      *AllLog = New(os.Stdout, LoggerInfo)
	globalMu sync.RWMutex
	writeMu  sync.Mutex
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
func SetOutput(out io.Writer, writerType string) { Default().SetOutput(out, writerType) }
func SetType(t LoggerType)                       { Default().SetType(t) }
func SetFlags(f int)                             { Default().SetFlags(f) }
func SetDepth(d int)                             { Default().SetDepth(d) }
func SetFormat(f Format)                         { Default().SetFormat(f) }

func Debug(v ...any)   { Default().Debug(v...) }
func Info(v ...any)    { Default().Info(v...) }
func Warn(v ...any)    { Default().Warn(v...) }
func Error(v ...any)   { Default().Error(v...) }
func Fatal(v ...any)   { Default().Fatal(v...) }
func Success(v ...any) { Default().Success(v...) }
func Notice(v ...any)  { Default().Notice(v...) }

func Debugf(f string, v ...any)                          { Default().Debugf(f, v...) }
func Infof(f string, v ...any)                           { Default().Infof(f, v...) }
func Warnf(f string, v ...any)                           { Default().Warnf(f, v...) }
func Errorf(f string, v ...any)                          { Default().Errorf(f, v...) }
func Fatalf(f string, v ...any)                          { Default().Fatalf(f, v...) }
func Successf(f string, v ...any)                        { Default().Successf(f, v...) }
func Noticef(f string, v ...any)                         { Default().Noticef(f, v...) }
func Customf(levelName string, f string, v ...any)       { Default().Customf(levelName, f, v...) }
func Streamf(f string, v ...any)                         { Default().Streamf(f, v...) }
func CustomStreamf(levelName string, f string, v ...any) { Default().CustomStreamf(levelName, f, v...) }

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
// default writer is: default; File writer is flog
func (l *AllLog) SetOutput(output io.Writer, writer string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if writer == "" || writer == "default" {
		l.slog.SetOutput(output)
		return
	}
	l.flog.SetOutput(output)
}

// SetDepth - Set depth to look for file. If 0 no filename will be listed in log
func (l *AllLog) SetDepth(depth int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	mask := log.Lshortfile | log.Llongfile
	if l.slog.Flags()&mask == 0 {
		fmt.Print("\nWARNING. YOU DO NOT HAVE A FLAG SPECIFIED IN slog INSTANCE THAT ENABLES depth SUPPORT.\n")
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

func (l *AllLog) SetFormat(f Format) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.format = f
}

func (l *AllLog) Format() Format {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.format
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
	message := fmt.Sprintf(format, v...)
	l.isStreaming = false
	fm, d := l.format, l.depth
	l.mu.Unlock()
	l.emit(fm, shallower(d), LoggerInfo, levelName, prefix, message, nil)
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

	if l.isStreaming && l.format == FormatColor {
		fmt.Print("\033[1A\033[2K") // mv1up & clear full
	}

	l.emit(l.format, shallower(l.depth), LoggerInfo, "stream", prefix, message, nil)

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

	if l.isStreaming && l.format == FormatColor {
		fmt.Print("\033[1A\033[2K")
	}

	l.emit(l.format, shallower(l.depth), LoggerInfo, levelName, prefix, message, nil)
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
	fm, d := l.format, l.depth
	l.mu.Unlock()

	var message string
	if len(v) > 0 && len(format) > 0 {
		message = fmt.Sprintf(format, v...)
	} else {
		message = fmt.Sprint(v...)
	}

	return l.emit(fm, d, tp, "", tp.toString(), message, nil)
}

func (l *AllLog) emit(fm Format, depth int, tp LoggerType, custom, colored, msg string, fields Fields) string {
	if fm == FormatLogfmt {
		level := tp.name()
		if custom != "" {
			level = strings.ToLower(custom)
		}
		var b strings.Builder
		b.WriteString("level=")
		b.WriteString(level)
		if depth > 0 {
			if _, file, line, ok := runtime.Caller(depth); ok {
				b.WriteString(" file=")
				b.WriteString(filepath.Base(file))
				b.WriteByte(':')
				b.WriteString(strconv.Itoa(line))
			}
		}
		b.WriteString(" msg=")
		b.WriteString(LogfmtValue(msg))
		b.WriteString(formatFields(fields, fm))
		out := b.String()
		writeMu.Lock()
		_, _ = io.WriteString(l.slog.Writer(), out+"\n")
		if l.flog != nil {
			_, _ = io.WriteString(l.flog.Writer(), out+"\n")
		}
		writeMu.Unlock()
		return out
	}

	prefix := colored
	if fm == FormatPlain {
		prefix = tp.plain()
		if custom != "" {
			prefix = "[" + strings.ToUpper(custom) + "] "
		}
	}
	out := prefix + msg + formatFields(fields, fm)
	_ = l.slog.Output(depth+1, out)
	if l.flog != nil {
		_ = l.flog.Output(depth+1, out)
	}
	return out
}

func shallower(depth int) int {
	if depth > 1 {
		return depth - 1
	}
	return depth
}

func LogfmtValue(v any) string {
	s := fmt.Sprint(v)
	if s == "" {
		return `""`
	}
	for _, r := range s {
		if r <= ' ' || r == '=' || r == '"' || r == 0x7f || r == '\uFFFD' {
			return strconv.Quote(s)
		}
	}
	return s
}

func formatFields(fields Fields, fm Format) string {
	if len(fields) == 0 {
		return ""
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		b.WriteByte(' ')
		switch fm {
		case FormatLogfmt:
			b.WriteString(k + "=" + LogfmtValue(fields[k]))
		case FormatPlain:
			b.WriteString(fmt.Sprintf("%s=%v", k, fields[k]))
		default:
			b.WriteString(fmt.Sprintf("%s%s%s=%v", ColorCyan, k, ColorReset, fields[k]))
		}
	}
	return b.String()
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

func (e *Entry) log(tp LoggerType, format string, v []any) {
	e.logger.mu.RLock()
	lvl, fm, d := e.logger.tp, e.logger.format, e.logger.depth
	e.logger.mu.RUnlock()
	if lvl > tp {
		return
	}

	var msg string
	if format == "" {
		msg = fmt.Sprint(v...)
	} else {
		msg = fmt.Sprintf(format, v...)
	}

	e.logger.emit(fm, shallower(d), tp, "", tp.toString(), msg, e.fields)
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

func (e *Entry) Info(v ...any)                    { e.log(LoggerInfo, "", v) }
func (e *Entry) Error(v ...any)                   { e.log(LoggerError, "", v) }
func (e *Entry) Debug(v ...any)                   { e.log(LoggerDebug, "", v) }
func (e *Entry) Warn(v ...any)                    { e.log(LoggerWarn, "", v) }
func (e *Entry) Success(v ...any)                 { e.log(LoggerSuccess, "", v) }
func (e *Entry) Infof(format string, v ...any)    { e.log(LoggerInfo, format, v) }
func (e *Entry) Errorf(format string, v ...any)   { e.log(LoggerError, format, v) }
func (e *Entry) Debugf(format string, v ...any)   { e.log(LoggerDebug, format, v) }
func (e *Entry) Warnf(format string, v ...any)    { e.log(LoggerWarn, format, v) }
func (e *Entry) Successf(format string, v ...any) { e.log(LoggerSuccess, format, v) }
