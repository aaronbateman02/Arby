package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

func New(level string) (*slog.Logger, error) {
	var l slog.Level
	switch level {
	case "debug":
		l = slog.LevelDebug
	case "info":
		l = slog.LevelInfo
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		return nil, fmt.Errorf("invalid log level: %s", level)
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: l,
	})), nil
}

func NewWriter(logger *slog.Logger, level slog.Level) io.Writer {
	r, w := io.Pipe()
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := r.Read(buf)
			if err != nil {
				return
			}
			logger.LogAttrs(nil, level, string(buf[:n]))
		}
	}()
	return w
}
