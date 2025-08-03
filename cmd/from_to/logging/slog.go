package logging

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func SetupSlog(isDebug, noColor bool, logFormat string) error {
	logHandlerOptions := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: isDebug,
	}
	if isDebug {
		logHandlerOptions.Level = slog.LevelDebug
	}

	var logHandler slog.Handler
	switch logFormat {
	case "json":
		logHandler = slog.NewJSONHandler(os.Stdout, logHandlerOptions)

	case "text":
		logHandler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      logHandlerOptions.Level,
			AddSource:  logHandlerOptions.AddSource,
			TimeFormat: time.DateTime,
			NoColor:    noColor,
		})

	default:
		return errors.New("invalid log format")
	}

	slog.SetDefault(slog.New(logHandler))

	return nil
}
