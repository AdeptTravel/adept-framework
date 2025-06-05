// internal/requestinfo/requestinfo.go
//
// Request metadata collector.
//
// Context
// -------
// This package provides inert structs (`RequestInfo`, `UA`, `Geo`) plus helper
// functions that enrich an incoming HTTP request with user-agent details,
// geolocation hints (MaxMind GeoLite 2), and a timestamp.  The structs carry
// **no** database handles or large buffers, so they are safe to log, JSON-encode,
// or attach to `context.Context`.
//
// Workflow
// --------
//  1. `InitGeo` is called at boot with the path to GeoLite2-City.  It opens a
//     singleton reader (`geoReader`) which is thread-safe for look-ups.
//  2. Middleware (not in this file) parses the User-Agent header with
//     `uasurfer.Parse`, resolves the client IP’s geo record, and stores a
//     `*RequestInfo` value in `context.Context` under an unexported key.
//  3. Handlers, Components, or Widgets call `requestinfo.FromContext()` to
//     retrieve the struct and render analytics attributes or template variables.

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
)

//
//  Struct definitions
//

// UA holds the parsed user-agent properties requested by Brandon.
type UA struct {
	Raw         string // Entire User-Agent header
	Browser     string // "Chrome", "Firefox", "Safari", etc.
	Version     string // "124.0.6367"
	OS          string // "macOS", "Windows", "Android", "iOS", etc.
	OSVersion   string // "14.5", "11", "10.0"
	Device      string // "Desktop", "Phone", "Tablet", "TV", ...
	Platform    string // "Mac", "Windows", "Linux", "iPad", ...
	IsBot       bool   // True if UA matches ~18 000 crawler signatures
	PrimaryLang string // First tag from Accept-Language ("en", "es", ...)
}

// Geo holds IP-based geolocation hints.
type Geo struct {
	IP         net.IP // Original client address (not X-Forwarded-For chain)
	CountryISO string // "US", "CA", "FR", ...
	City       string // "Chicago", "Paris", ...
}

// RequestInfo is attached to the project’s Context type and is therefore
// visible to Components, Widgets, and templates.
type RequestInfo struct {
	UA        UA
	Geo       Geo
	URL       *url.URL // Pointer copy, safe to dereference read-only
	Timestamp time.Time
}

//
//  Package-level state
//

// geoReader is a singleton MaxMind handle.  It is safe for concurrent reads.
var geoReader *geoip2.Reader

// InitGeo opens the GeoLite2-City database at startup.  Call from main().
func InitGeo(dbPath string) {
	var err error
	geoReader, err = geoip2.Open(dbPath)
	if err != nil {
		panic("requestinfo: cannot open GeoLite2 DB: " + err.Error())
	}
}

//
//  Public helper: FromContext
//

type ctxKey struct{}

func FromContext(ctx context.Context) *RequestInfo {
	v, _ := ctx.Value(ctxKey{}).(*RequestInfo)
	return v
}

//
//  Internal helpers
//

// parseUA converts a raw header into our UA struct using uasurfer.
func parseUA(uaHeader, acceptLang string) UA {
	u := uasurfer.Parse(uaHeader)

	// Browser family
	br := strings.TrimPrefix(u.Browser.Name.String(), "Browser")

	// Browser version "major.minor.patch"
	brVer := trimVersion(u.Browser.Version)

	// OS name and version
	osName := strings.TrimPrefix(u.OS.Name.String(), "OS")
	if osName == "MacOSX" {
		osName = "macOS"
	}
	osVer := trimVersion(u.OS.Version)

	// Device class
	device := deviceTypeToString(u.DeviceType)

	// Platform string via OS.Platform (uasurfer ≈2.x)
	platform := strings.TrimPrefix(u.OS.Platform.String(), "Platform")

	return UA{
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
}

// trimVersion builds "major.minor.patch" and removes trailing ".0".
func trimVersion(v uasurfer.Version) string {
	out := strings.TrimSuffix(
		strings.TrimSuffix(
			strings.TrimSuffix(
				strings.Join([]string{
					intToStr(uint64(v.Major)),
					intToStr(uint64(v.Minor)),
					intToStr(uint64(v.Patch)),
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

// intToStr converts uint64 to string.
func intToStr(i uint64) string {
	return strconv.FormatUint(i, 10)
}

// deviceTypeToString maps uasurfer.DeviceType to a user-friendly string.
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

// primaryLang extracts the first language subtag before any ";q=" rule.
func primaryLang(al string) string {
	if al == "" {
		return ""
	}
	parts := strings.Split(al, ",")
	tag := strings.TrimSpace(parts[0])
	if i := strings.Index(tag, ";"); i != -1 {
		tag = tag[:i]
	}
	return strings.ToLower(tag)
}

// lookupGeo returns best-effort Geo data using the global reader.
func lookupGeo(ip net.IP) Geo {
	if geoReader == nil || ip == nil {
		return Geo{IP: ip}
	}
	rec, err := geoReader.City(ip)
	if err != nil {
		return Geo{IP: ip}
	}
	return Geo{
		IP:         ip,
		CountryISO: rec.Country.IsoCode,
		City:       rec.City.Names["en"],
	}
}
