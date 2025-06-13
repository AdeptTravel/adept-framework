// internal/requestinfo/requestinfo.go
//
// Request metadata collector.
//
// Context
// --------
// This package enriches an incoming HTTP request with UA details parsed by
// `internal/ua`, IP‑based geolocation, the canonical URL, and a UTC
// timestamp.  All helpers emit DEBUG‑level Zap spans so developers can
// trace GeoIP look‑ups when `ZAP_LEVEL=debug`.
//
// Instrumentation
// ---------------
//   - `InitGeo`   – INFO on success, FATAL on failure to open the DB.
//   - `lookupGeo` – DEBUG with IP, country ISO, and city (cache hits only).
//
// Notes
// -----
//   - All structs are inert (no DB handles) and safe to JSON‑encode.
package requestinfo

import (
	"context"
	"net"
	"net/url"
	"time"

	"github.com/oschwald/geoip2-golang"
	"github.com/yanizio/adept/internal/ua"
	"go.uber.org/zap"
)

/*──────────────────────────── struct definitions ───────────────────────────*/

type Geo struct {
	IP         net.IP
	CountryISO string
	City       string
}

type RequestInfo struct {
	UA        ua.Info
	Geo       Geo
	URL       *url.URL
	Timestamp time.Time
}

/*──────────────────────────── package‑level state ──────────────────────────*/

var geoReader *geoip2.Reader // singleton, safe for concurrent reads

// InitGeo opens the GeoLite2‑City DB at startup.
func InitGeo(dbPath string) {
	var err error
	geoReader, err = geoip2.Open(dbPath)
	if err != nil {
		zap.S().Fatal("geo db open failed", zap.String("path", dbPath), zap.Error(err))
	}
	zap.S().Infow("geo db ready", "path", dbPath)
}

/*────────────────────────── context accessor ───────────────────────────────*/

type ctxKey struct{}

func FromContext(ctx context.Context) *RequestInfo {
	v, _ := ctx.Value(ctxKey{}).(*RequestInfo)
	return v
}

/*──────────────────────────── internal helpers ─────────────────────────────*/

// lookupGeo returns best‑effort Geo data.
func lookupGeo(ip net.IP) Geo {
	if geoReader == nil || ip == nil {
		return Geo{IP: ip}
	}
	rec, err := geoReader.City(ip)
	if err != nil {
		return Geo{IP: ip}
	}
	geo := Geo{
		IP:         ip,
		CountryISO: rec.Country.IsoCode,
		City:       rec.City.Names["en"],
	}
	zap.S().Debugw("geo lookup", "ip", geo.IP, "country", geo.CountryISO, "city", geo.City)
	return geo
}
