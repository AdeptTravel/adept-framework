// internal/ua/parse.go
//
// Wrapper around github.com/mssola/user_agent with extra logic so modern
// macOS UAs like
//
//	"Macintosh; Intel Mac OS X 10_15_7"
//
// map to OS = "Mac OS X", OSVersion = "10.15.7" instead of the misleading
// default "Intel".
package ua

import (
	"strings"

	"github.com/mssola/user_agent"
)

type Info struct {
	Browser   string
	Version   string
	OS        string
	OSVersion string
	Device    string // "Desktop" or "Mobile"
	Platform  string // ua.Platform()  ("Macintosh", "X11", etc.)
	IsBot     bool
}

// Parse converts a raw User-Agent header into Info.
func Parse(raw string) Info {
	ua := user_agent.New(raw)

	browser, version := ua.Browser()
	osFull := ua.OS() // examples: "Intel Mac OS X 10_15_7", "Windows 10"

	// Canonicalise macOS strings: drop "Intel " or "PPC ".
	osFull = strings.TrimPrefix(osFull, "Intel ")
	osFull = strings.TrimPrefix(osFull, "PPC ")

	// Split first token (OS family) from remainder (version string).
	osParts := strings.SplitN(osFull, " ", 2)

	device := "Desktop"
	if ua.Mobile() {
		device = "Mobile"
	}

	info := Info{
		Browser:  browser,
		Version:  version,
		OS:       osParts[0],
		Device:   device,
		Platform: ua.Platform(),
		IsBot:    ua.Bot(),
	}

	// Normalise version: convert underscores to dots (macOS) and trim spaces.
	if len(osParts) == 2 {
		info.OSVersion = strings.ReplaceAll(strings.TrimSpace(osParts[1]), "_", ".")
	}
	return info
}
