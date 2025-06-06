// internal/requestinfo/requestinfo.go
//
// Request metadata collector.
//
/*
Context
--------
This package enriches an incoming HTTP request with user-agent details,
IP-based geolocation hints, canonical URL, and a UTC timestamp.  All
helpers emit DEBUG-level Zap spans so developers can trace UA parsing and
GeoIP look-ups when `ZAP_LEVEL=debug`.

Instrumentation
---------------
  • `InitGeo`  – INFO on success, FATAL on failure to open the MaxMind DB.
  • `parseUA`  – DEBUG with browser, device, bot flag.
  • `lookupGeo`– DEBUG with IP, country ISO, city (hits only).

Notes
-----
  • All structs are inert (no DB handles) and safe to JSON-encode.
  • Oxford commas, two spaces after periods.  No em dash.
*/
package requestinfo

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/avct/uasurfer"
	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"
)

/*──────────────────────────── struct definitions ───────────────────────────*/

type UA struct {
	Raw         string // full User-Agent header
	Browser     string // "Chrome", "Firefox", …
	Version     string // "124.0.6367"
	OS          string // "macOS", "Windows", …
	OSVersion   string // "14.5", "11", …
	Device      string // "Desktop", "Phone", …
	Platform    string // "Mac", "Windows", …
	IsBot       bool   // true if crawler
	PrimaryLang string // first tag from Accept-Language
}

type Geo struct {
	IP         net.IP
	CountryISO string
	City       string
}

type RequestInfo struct {
	UA        UA
	Geo       Geo
	URL       *url.URL
	Timestamp time.Time
}

/*──────────────────────────── package-level state ──────────────────────────*/

var geoReader *geoip2.Reader // singleton, safe for concurrent reads

// InitGeo opens the GeoLite2-City DB at startup.
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

// parseUA converts a raw header into UA struct.
func parseUA(uaHeader, acceptLang string) UA {
	u := uasurfer.Parse(uaHeader)

	br := strings.TrimPrefix(u.Browser.Name.String(), "Browser")
	brVer := trimVersion(u.Browser.Version)

	osName := strings.TrimPrefix(u.OS.Name.String(), "OS")
	if osName == "MacOSX" {
		osName = "macOS"
	}
	osVer := trimVersion(u.OS.Version)

	device := deviceTypeToString(u.DeviceType)
	platform := strings.TrimPrefix(u.OS.Platform.String(), "Platform")

	ua := UA{
		Raw:         uaHeader,
		Browser:     br,
		Version:     brVer,
		OS:          osName,
		OSVersion:   osVer,
		Device:      device,
		Platform:    platform,
		IsBot:       u.IsBot(),
		PrimaryLang: primaryLang(acceptLang),
	}

	zap.S().Debugw("ua parsed",
		"browser", ua.Browser,
		"device", ua.Device,
		"bot", ua.IsBot,
	)
	return ua
}

// lookupGeo returns best-effort Geo data.
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

/*────────────────────────── utility functions ─────────────────────────────*/

func trimVersion(v uasurfer.Version) string {
	out := strings.TrimSuffix(
		strings.TrimSuffix(
			strings.TrimSuffix(
				strings.Join([]string{
					strconv.Itoa(v.Major),
					strconv.Itoa(v.Minor),
					strconv.Itoa(v.Patch),
				}, "."),
				".0"),
			".0"),
		".0",
	)
	if out == "" {
		return "0"
	}
	return out
}

func deviceTypeToString(dt uasurfer.DeviceType) string {
	switch dt {
	case uasurfer.DeviceComputer:
		return "Desktop"
	case uasurfer.DevicePhone:
		return "Phone"
	case uasurfer.DeviceTablet:
		return "Tablet"
	case uasurfer.DeviceConsole:
		return "Console"
	case uasurfer.DeviceWearable:
		return "Wearable"
	case uasurfer.DeviceTV:
		return "TV"
	default:
		return "Unknown"
	}
}

func primaryLang(al string) string {
	if al == "" {
		return ""
	}
	tag := strings.TrimSpace(strings.Split(al, ",")[0])
	if i := strings.Index(tag, ";"); i != -1 {
		tag = tag[:i]
	}
	return strings.ToLower(tag)
}
