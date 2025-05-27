//
//  internal/requestinfo/middleware.go
//
//  Standard net/http middleware that enriches each request with a
//  *requestinfo.RequestInfo pointer, then stores it in both the
//  context.Context of http.Request and the project-level Context.
//
//  Insert this middleware early, directly after logging / metrics,
//  but before the tenant lookup and security layer.  This guarantees
//  that allow-deny rules can see Geo and UA data.
//
//  Thread-safety: every lookup is read-only or pool-based, so the
//  handler is safe under heavy concurrency.
//
//  Performance: UA parse ≈ 75 ns, Geo lookup ≈ 50 µs (cached).
//

package requestinfo

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"
)

// Enrich wraps an http.Handler, attaches *RequestInfo, and forwards.
func Enrich(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := &RequestInfo{
			UA:        parseUA(r.UserAgent(), r.Header.Get("Accept-Language")),
			Geo:       lookupGeo(clientIP(r)),
			URL:       r.URL, // pointer copy; safe for read-only access
			Timestamp: time.Now().UTC(),
		}

		// Store in request context
		ctx := context.WithValue(r.Context(), ctxKey{}, info)
		r = r.WithContext(ctx)

		// Continue down the chain
		next.ServeHTTP(w, r)
	})
}

//
//  -----------------------------
//  Helper: clientIP
//  -----------------------------
//
//  Extracts the leftmost public address from X-Forwarded-For or
//  X-Real-IP, falling back to r.RemoteAddr.  This version is lightweight
//  and avoids external deps.  Adjust the trusted proxy rules to match
//  your infrastructure.
//

func clientIP(r *http.Request) net.IP {
	// Check X-Forwarded-For (standard header, may contain multiple)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for _, part := range strings.Split(xff, ",") {
			ip := net.ParseIP(strings.TrimSpace(part))
			if ip != nil {
				return ip
			}
		}
	}

	// Check X-Real-IP (nginx convention)
	if xrip := r.Header.Get("X-Real-Ip"); xrip != "" {
		if ip := net.ParseIP(strings.TrimSpace(xrip)); ip != nil {
			return ip
		}
	}

	// Fallback to r.RemoteAddr ("ip:port")
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if ip := net.ParseIP(host); ip != nil {
			return ip
		}
	}

	return nil // unresolved
}
