package log

import (
	"bytes"
	"strings"
	"testing"
)

func TestInfoOutputsMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	// redirect package logger output to our buffer
	new(AllLog).SetOutput(buf)
	// don't include file info. PS: need to manually set L* log flag
	//new(AllLog).SetDepth(0)
	// allow all levels (assumes LoggerDebug is the lowest value)
	new(AllLog).SetType(LoggerDebug)

	Info("hello %s", "world")

	got := buf.String()
	if !strings.Contains(got, "hello world") {
		t.Fatalf("expected log to contain message %q, got %q", "hello world", got)
	}
}

func TestInfoSuppressedWhenLevelHigher(t *testing.T) {
	buf := &bytes.Buffer{}
	new(AllLog).SetOutput(buf)
	//new(AllLog).SetDepth(0)
	// set logger to a higher level so Info should be suppressed; this assumes
	// LoggerError > LoggerInfo (CreatePerCall checks `if logger.tp > tp { return "" }`).
	new(AllLog).SetType(LoggerError)

	Info("this should not appear")
	if got := buf.String(); got != "" {
		t.Fatalf("expected no output when logger level is higher, got %q", got)
	}
}

func TestFatalPanicsWithMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	new(AllLog).SetOutput(buf)
	//new(AllLog).SetDepth(0)
	new(AllLog).SetType(LoggerDebug)

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected Fatal to panic, but it did not")
		} else {
			// ensure the panic value contains our message
			if s, ok := r.(string); ok {
				if !strings.Contains(s, "fatal test") {
					t.Fatalf("panic value does not contain expected message: %q", s)
				}
			} else {
				t.Fatalf("panic value has unexpected type: %T", r)
			}
		}
	}()

	Fatal("fatal %s", "test")
}
