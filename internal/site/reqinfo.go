package site

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/AdeptTravel/adept-framework/internal/geo"
)

// ReqInfo is rendered by the debug handler.
type ReqInfo struct {
	UserAgent string
	IP        string
	Host      string
	Path      string
	Time      time.Time
	Geo       *Geo // nil when lookup fails or DB not configured
}

// Geo is a friendly subset of the MaxMind record.
type Geo struct {
	CountryISO string
	Country    string
	City       string
	Lat, Lon   float64
}

// Extract builds ReqInfo from the request and optional GeoIP DB.
func Extract(r *http.Request, devHost string, geoDB *geo.DB) ReqInfo {
	// --- client IP ---
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		ip = strings.TrimSpace(strings.Split(xf, ",")[0])
	}

	// --- host ---
	host := r.Host
	if devHost != "" {
		host = devHost
	}

	// --- Geo lookup ---
	var g *Geo
	if geoDB != nil {
		if loc, err := geoDB.Lookup(ip); err == nil {
			g = &Geo{
				CountryISO: loc.CountryISO,
				Country:    loc.Country,
				City:       loc.City,
				Lat:        loc.Lat,
				Lon:        loc.Lon,
			}
		}
	}

	return ReqInfo{
		UserAgent: r.UserAgent(),
		IP:        ip,
		Host:      host,
		Path:      r.URL.Path,
		Time:      time.Now(),
		Geo:       g,
	}
}
