package logger

import (
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"kaffein/auth-service/config"
)

func New(cfg config.LoggerConfig) zerolog.Logger {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	var output io.Writer = os.Stdout
	if strings.EqualFold(cfg.Format, "console") {
		output = zerolog.ConsoleWriter{Out: os.Stdout}
	}
	return zerolog.New(output).Level(level).With().Timestamp().Caller().Logger()
}
