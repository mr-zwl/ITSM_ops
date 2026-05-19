package logger

import "testing"

func TestNew(t *testing.T) {
	for _, level := range []string{"debug", "info", "warn", "error", "unknown"} {
		l := New(level)
		if l == nil {
			t.Errorf("New(%q) returned nil", level)
		}
	}
}
