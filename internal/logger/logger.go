// internal/logger/logger.go
//
// Structured JSON logger (Zap + Lumberjack).
//
// Context
// -------
// Adept writes lifecycle and error events to one JSON log per day under
// `<root>/logs/YYYY-MM-DD.log`.  When running in an interactive TTY we tee
// the same events, colorized, to stdout.  Rotation, compression, and
// retention are handled by Lumberjack; no external log-rotate job is
// required.
//
// Usage
// -----
//
//	log, err := logger.New(cfg.Paths.Root, runningInTTY())
//	if err != nil { … }
//	log.Infow("tenant online", "tenant", host)
//
// Notes
// -----
// • Zap core uses ISO-8601 timestamps and lowercase levels.
// • Errors are written to the same sink via `ErrorOutput`.
// • Oxford commas, two spaces after periods.
package logger

import (
	"os"
	"path/filepath"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New returns a *zap.SugaredLogger that writes JSON to /logs/YYYY-MM-DD.log.
// When tee == true, a colored console core is also attached.  The logger
// is installed as the process-wide default via zap.ReplaceGlobals.
func New(rootDir string, tee bool) (*zap.SugaredLogger, error) {
	logDir := filepath.Join(rootDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}

	fileName := time.Now().Format("2006-01-02") + ".log"
	fileSink := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, fileName),
		MaxSize:    50, // MB
		MaxBackups: 7,  // keep last seven files
		MaxAge:     14, // days
		Compress:   true,
	}

	encCfg := zapcore.EncoderConfig{
		TimeKey:      "ts",
		LevelKey:     "level",
		MessageKey:   "msg",
		CallerKey:    "caller",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeLevel:  zapcore.LowercaseLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}
	jsonCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encCfg),
		zapcore.AddSync(fileSink),
		zap.InfoLevel,
	)

	var cores []zapcore.Core
	cores = append(cores, jsonCore)

	if tee {
		consoleCore := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encCfg),
			zapcore.AddSync(os.Stdout),
			zap.InfoLevel,
		)
		cores = append(cores, consoleCore)
	}

	z := zap.New(
		zapcore.NewTee(cores...),
		zap.ErrorOutput(zapcore.AddSync(fileSink)),
	).Sugar()

	// Make this the global logger so zap.L() works everywhere after startup.
	zap.ReplaceGlobals(z.Desugar())

	z.Infow("logger online", "tee", tee)
	return z, nil
}
