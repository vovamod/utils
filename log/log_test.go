package log

import (
	"bytes"
	"log"
	"os"
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

// Global logger instance
func TestGlobalLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	customLog := New(buf, LoggerDebug)
	Replace(customLog)

	t.Run("Global Standard Logging", func(t *testing.T) {
		buf.Reset()
		Info("global info message")

		output := buf.String()
		if !strings.Contains(output, "global info message") {
			t.Errorf("Expected output to contain message, got: %q", output)
		}
	})

	t.Run("Global Formatted Logging", func(t *testing.T) {
		buf.Reset()
		Successf("task %d completed", 1)

		output := buf.String()
		if !strings.Contains(output, "task 1 completed") {
			t.Errorf("Expected formatted output, got: %q", output)
		}
	})

	t.Run("Global Level Filtering", func(t *testing.T) {
		buf.Reset()
		SetType(LoggerError)

		Debug("this should not appear")

		if buf.Len() > 0 {
			t.Errorf("Expected no output for Debug at Error level, got: %q", buf.String())
		}

		SetType(LoggerDebug)
	})
}

func TestGlobalStructuredLogging(t *testing.T) {
	buf := &bytes.Buffer{}
	Replace(New(buf, LoggerDebug))

	t.Run("Global WithField", func(t *testing.T) {
		buf.Reset()
		WithField("component", "api").Info("request received")

		output := buf.String()
		if !strings.Contains(output, "component") || !strings.Contains(output, "api") || !strings.Contains(output, "request received") {
			t.Errorf("Structured log missing field or message. Got: %q", output)
		}
	})

	t.Run("Global WithFields", func(t *testing.T) {
		buf.Reset()
		WithFields(Fields{
			"user_id": 42,
			"role":    "admin",
		}).Warn("unauthorized access attempt")

		output := buf.String()
		keys := []string{"user_id", "42", "role", "admin", "unauthorized access attempt"}
		for _, key := range keys {
			if !strings.Contains(output, key) {
				t.Errorf("Expected output to contain %s, but it didn't. Got: %q", key, output)
			}
		}
	})
}

func TestGlobalConcurrency(t *testing.T) {
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			Info("logging...")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			Replace(New(os.Stdout, LoggerInfo))
		}
		done <- true
	}()

	<-done
	<-done
}
