package logging

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
)

var levelMap = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func ParseLevel(raw string) (slog.Level, error) {
	level, ok := levelMap[strings.ToLower(strings.TrimSpace(raw))]
	if !ok {
		return slog.LevelInfo, fmt.Errorf("invalid log level %q", raw)
	}
	return level, nil
}

func NewJSONLogger(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level}))
}

func SafeErr(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if len(msg) > 256 {
		return msg[:256]
	}
	return msg
}
