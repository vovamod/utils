package log

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func newTestLogger(level LoggerType) (*AllLog, *bytes.Buffer) {
	buf := &bytes.Buffer{}

	l := &AllLog{
		slog:   log.New(buf, "", 0),
		tp:     level,
		depth:  0,
		exitFn: func(int) {},
	}

	return l, buf
}

func TestInfoLog(t *testing.T) {
	l, buf := newTestLogger(LoggerInfo)

	l.Info("hello world")

	out := buf.String()
	if !strings.Contains(out, "hello world") {
		t.Fatalf("expected log output, got: %s", out)
	}
}

func TestLogLevelFiltering(t *testing.T) {
	l, buf := newTestLogger(LoggerWarn)

	l.Debug("debug should not appear")
	l.Info("info should not appear")
	l.Warn("warn should appear")

	out := buf.String()
	log.Print(out)

	if strings.Contains(out, "debug") {
		t.Fatal("debug should be filtered")
	}
	if strings.Contains(out, "info") {
		t.Fatal("info should be filtered")
	}
	if !strings.Contains(out, "warn") {
		t.Fatal("warn should appear")
	}
}

func TestWithField(t *testing.T) {
	l, buf := newTestLogger(LoggerInfo)

	l.WithField("user", 123).Info("event")

	out := buf.String()

	if !strings.Contains(out, "user") || !strings.Contains(out, "123") {
		t.Fatalf("expected field output, got: %s", out)
	}
}

func TestWithFieldsMultiple(t *testing.T) {
	l, buf := newTestLogger(LoggerInfo)

	l.WithFields(Fields{
		"a": 1,
		"b": 2,
	}).Infof("multi")

	out := buf.String()

	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Fatalf("expected multiple fields, got: %s", out)
	}
}

func TestFieldChaining(t *testing.T) {
	l, buf := newTestLogger(LoggerInfo)

	l.WithField("a", 1).
		WithField("b", 2).
		Infof("chain")

	out := buf.String()

	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Fatalf("expected chained fields, got: %s", out)
	}
}

func TestFieldOverwrite(t *testing.T) {
	l, buf := newTestLogger(LoggerInfo)

	l.WithField("key", 1).
		WithField("key", 2).
		Infof("overwrite")

	out := buf.String()

	if !strings.Contains(out, "key") {
		t.Fatalf("expected key in output: %s", out)
	}
	if !strings.Contains(out, "2") {
		t.Fatalf("expected overwritten value 2, got: %s", out)
	}
}

func TestFatalDoesNotExit(t *testing.T) {
	l, buf := newTestLogger(LoggerInfo)

	called := false
	l.exitFn = func(code int) {
		called = true
	}

	l.Fatal("critical error")

	if !called {
		t.Fatal("expected exitFn to be called")
	}

	if !strings.Contains(buf.String(), "critical error") {
		t.Fatal("expected fatal log output")
	}
}

func TestEntryIsIndependent(t *testing.T) {
	l, buf := newTestLogger(LoggerInfo)

	e1 := l.WithField("a", 1)
	e2 := e1.WithField("b", 2)

	e1.Infof("first")
	e2.Infof("second")

	out := buf.String()

	if !strings.Contains(out, "first") || !strings.Contains(out, "second") {
		t.Fatalf("expected both logs, got: %s", out)
	}
}
