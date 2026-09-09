package logger

import (
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"kaffein/product-service/config"
)

var (
	once    sync.Once
	znLog   zerolog.Logger
	initErr error
)

func Get() zerolog.Logger {
	return znLog
}

func Init(cfg *config.Config, hooks ...zerolog.Hook) error {
	once.Do(func() {
		znLog, initErr = newLogger(cfg, hooks...)
		if initErr != nil {
			znLog = zerolog.Nop()
			return
		}

		log.Logger = znLog
	})
	return initErr
}

func New(cfg *config.Config, hooks ...zerolog.Hook) {
	Init(cfg, hooks...)
}

func NewWithWriter(level string, w io.Writer, hooks ...zerolog.Hook) zerolog.Logger {
	lvl := parseLevel(level)
	zerolog.SetGlobalLevel(lvl)

	logger := zerolog.New(w).
		With().
		Timestamp().
		Logger().
		Level(lvl)

	for _, h := range hooks {
		logger = logger.Hook(h)
	}

	return logger
}

func newLogger(cfg *config.Config, hooks ...zerolog.Hook) (zerolog.Logger, error) {
	level := parseLevel(cfg.Logger.Level)
	zerolog.SetGlobalLevel(level)

	output := buildOutput(cfg)

	logger := zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger().
		Level(level)

	for _, h := range hooks {
		logger = logger.Hook(h)
	}

	return logger, nil
}

func buildOutput(cfg *config.Config) io.Writer {
	var writers []io.Writer

	if cfg.Logger.Stdout {
		writers = append(writers, wrapWriter(os.Stdout, cfg.Logger.Format))
	}

	if cfg.Logger.Path != "" {
		if f, err := openLogFile(cfg.Logger.Path); err != nil {
			os.Stderr.WriteString("logger: failed to open log file: " + err.Error() + "\n")
		} else {
			writers = append(writers, wrapWriter(f, cfg.Logger.Format))
		}
	}

	return combineWriters(writers)
}

func openLogFile(path string) (*os.File, error) {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
}

func wrapWriter(w io.Writer, format string) io.Writer {
	if format == "console" {
		return zerolog.ConsoleWriter{Out: w, NoColor: w != os.Stdout}
	}
	return w
}

func combineWriters(writers []io.Writer) io.Writer {
	switch len(writers) {
	case 0:
		return os.Stderr
	case 1:
		return writers[0]
	default:
		return io.MultiWriter(writers...)
	}
}

func parseLevel(level string) zerolog.Level {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.InfoLevel
	}
	return lvl
}

func GetBuildInfo() (revision, goVersion string) {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "", ""
	}

	for _, v := range buildInfo.Settings {
		if v.Key == "vcs.revision" {
			revision = v.Value
		}
	}

	return revision, buildInfo.GoVersion
}

func ConsoleWriter() zerolog.ConsoleWriter {
	return zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}
}

func MultiLevelWriter(writers ...io.Writer) io.Writer {
	return zerolog.MultiLevelWriter(writers...)
}
