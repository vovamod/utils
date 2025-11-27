package log

import (
	"log"
)

// LoggerType - for validation
type LoggerType int

// ANSI color codes
const (
	ColorReset        = "\033[0m " // extra space HERE in order to not write spaces to each of log entries
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
)

// Level
const (
	LoggerDebug LoggerType = iota
	LoggerInfo
	LoggerWarn
	LoggerError
	LoggerFatal
	LoggerSuccess
	LoggerNotice
)

var lstr = []string{
	ColorBlue + "[DEBUG]" + ColorReset,
	ColorBrightGreen + "[INFO]" + ColorReset,
	ColorBrightYellow + "[WARN]" + ColorReset,
	ColorBrightRed + "[ERROR]" + ColorReset,
	ColorBgDarkRed + "[FATAL]" + ColorReset,
	ColorGreen + "[SUCCESS]" + ColorReset,
	ColorYellow + "[NOTICE]" + ColorReset,
}

func (lT LoggerType) toString() string {
	if lT >= LoggerDebug && lT <= LoggerNotice {
		return lstr[lT]
	}
	return "[UNKNOWN]"
}

func (lT LoggerType) IsValid() bool {
	switch lT {
	case LoggerDebug, LoggerInfo, LoggerWarn, LoggerError, LoggerFatal, LoggerSuccess, LoggerNotice:
		return true
	default:
		return false
	}
}

// AllLog - represents logger and it's interface with params
type AllLog struct {
	slog   *log.Logger
	depth  int        // for debug only
	tp     LoggerType // log type
	Logger            // Logger interface (used to call func-s)
}
