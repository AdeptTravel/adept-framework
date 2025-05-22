// Package logger builds a *log.Logger that always writes to a dated log file
// under /log, and optionally tees output to stdout when running in an
// interactive TTY.  The helper is intentionally minimal so it can be reused
// by CLI tools and background workers.
package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// New returns a logger that writes to /log/YYYY-MM-DD.log relative to the
// supplied root directory.  When tee is true, the logger also writes to
// stdout, making local development easier.
func New(rootDir string, tee bool) (*log.Logger, error) {
	logDir := filepath.Join(rootDir, "log")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}

	fileName := time.Now().Format("2006-01-02") + ".log"
	f, err := os.OpenFile(filepath.Join(logDir, fileName),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	var w io.Writer = f
	if tee {
		w = io.MultiWriter(os.Stdout, f)
	}

	l := log.New(w, "", log.LstdFlags|log.Lshortfile)
	l.Printf("logger online (tee=%v)", tee)
	return l, nil
}
