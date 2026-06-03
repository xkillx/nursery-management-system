package logging

import (
	"log/slog"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  slog.Level
		ok    bool
	}{
		{"debug", slog.LevelDebug, true},
		{"info", slog.LevelInfo, true},
		{"warn", slog.LevelWarn, true},
		{"error", slog.LevelError, true},
		{"DEBUG", slog.LevelDebug, true},
		{"Info", slog.LevelInfo, true},
		{"  warn  ", slog.LevelWarn, true},
		{"verbose", slog.LevelInfo, false},
		{"", slog.LevelInfo, false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := ParseLevel(tc.input)
			if tc.ok {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tc.want {
					t.Fatalf("expected %v, got %v", tc.want, got)
				}
			} else {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			}
		})
	}
}

func TestSafeErr(t *testing.T) {
	t.Run("nil error returns empty", func(t *testing.T) {
		if got := SafeErr(nil); got != "" {
			t.Fatalf("expected empty, got %q", got)
		}
	})

	t.Run("short error preserved", func(t *testing.T) {
		got := SafeErr(someErr("connection refused"))
		if got != "connection refused" {
			t.Fatalf("expected 'connection refused', got %q", got)
		}
	})

	t.Run("long error truncated", func(t *testing.T) {
		longMsg := make([]byte, 512)
		for i := range longMsg {
			longMsg[i] = 'a'
		}
		got := SafeErr(someErr(string(longMsg)))
		if len(got) != 256 {
			t.Fatalf("expected 256 chars, got %d", len(got))
		}
	})
}

type someErr string

func (e someErr) Error() string { return string(e) }
