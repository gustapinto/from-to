package logging

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

const (
	_logFormatJson = "json"
	_logFormatText = "text"
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
	case _logFormatJson:
		logHandler = slog.NewJSONHandler(os.Stdout, logHandlerOptions)

	case _logFormatText:
		logHandler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      logHandlerOptions.Level,
			AddSource:  logHandlerOptions.AddSource,
			TimeFormat: time.DateTime,
			NoColor:    noColor,
		})

	default:
		return fmt.Errorf("invalid log format [%s]", logFormat)
	}

	slog.SetDefault(slog.New(logHandler))

	return nil
}
