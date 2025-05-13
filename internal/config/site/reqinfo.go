package site

import (
    "net"
    "net/http"
    "strings"
    "time"
)

// ReqInfo captures per-request details for debugging or logging.
type ReqInfo struct {
    UserAgent string
    IP        string
    Host      string
    Path      string
    Time      time.Time
}

// Extract populates ReqInfo from an *http.Request.
// devHost overrides r.Host when supplied (useful for local testing).
func Extract(r *http.Request, devHost string) ReqInfo {
    ip, _, _ := net.SplitHostPort(r.RemoteAddr)
    if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
        ip = strings.TrimSpace(strings.Split(xf, ",")[0])
    }

    host := r.Host
    if devHost != "" {
        host = devHost
    }

    return ReqInfo{
        UserAgent: r.UserAgent(),
        IP:        ip,
        Host:      host,
        Path:      r.URL.Path,
        Time:      time.Now(),
    }
}
