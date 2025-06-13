// internal/ua/ua.go
//
// User‑Agent parsing helpers.
//
// This wrapper isolates the third‑party `github.com/avct/uasurfer` API so
// the rest of the codebase never sees its enums or structs.  If we ever
// swap parsers again, only this file changes.
package ua

import (
	"fmt"
	"strconv"

	surfer "github.com/avct/uasurfer"
)

// Info carries the UA attributes used by middleware, Components, and
// templates.  Field names mirror the previous parser, so downstream code
// compiles unchanged.
//
// Example (Chrome on macOS):
//
//	Browser   "Chrome"
//	Version   "125.0.6422"
//	OS        "Mac OS X"
//	OSVersion "14.4"
//	Device    "Desktop"
//	Platform  "Macintosh"
//	IsBot     false
//	Raw       "Mozilla/5.0 (Macintosh;…"
//
// Device will be one of: "Desktop", "Mobile", "Tablet", or "Other".
type Info struct {
	Browser   string
	Version   string
	OS        string
	OSVersion string
	Device    string
	Platform  string
	IsBot     bool
	Raw       string
}

// Parse converts a raw header into an Info struct.  After the first call
// the underlying library reuses internal buffers, so Parse allocates only
// on rarely‑seen strings.
func Parse(raw string) Info {
	ua := surfer.Parse(raw)

	info := Info{
		Browser:   ua.Browser.Name.String(),
		Version:   versionToString(ua.Browser.Version),
		OS:        ua.OS.Name.String(),
		OSVersion: versionToString(ua.OS.Version),
		Platform:  ua.OS.Platform.String(),
		IsBot:     ua.IsBot(),
		Raw:       raw,
	}

	switch ua.DeviceType {
	case surfer.DeviceComputer:
		info.Device = "Desktop"
	case surfer.DeviceTablet:
		info.Device = "Tablet"
	case surfer.DevicePhone, surfer.DeviceWearable:
		info.Device = "Mobile"
	default:
		info.Device = "Other"
	}

	return info
}

// versionToString renders a semantic version in dotted form while trimming
// trailing zeros, e.g. 17.0.0 → "17", 17.3.0 → "17.3", 17.3.1 → "17.3.1".
func versionToString(v surfer.Version) string {
	if v.Major == 0 && v.Minor == 0 && v.Patch == 0 {
		return ""
	}
	if v.Patch != 0 {
		return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	}
	if v.Minor != 0 {
		return fmt.Sprintf("%d.%d", v.Major, v.Minor)
	}
	return strconv.Itoa(int(v.Major))
}
