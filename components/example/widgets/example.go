// components/example/widgets/example.go
//
// Example widget — outputs UA/IP/Geo info in a snippet.
package widgets

import (
	"github.com/yanizio/adept/internal/tenant"
	"github.com/yanizio/adept/internal/view"
	"github.com/yanizio/adept/internal/widget"
)

// compile-time assertion
var _ widget.Widget = (*Widget)(nil)

// Widget implements widget.Widget.
type Widget struct{}

func (w *Widget) ID() string { return "example/example" }

// Render converts ctx to *tenant.Context, builds data, and returns the
// rendered HTML string plus cache policy.
func (w *Widget) Render(ctx any, _ map[string]any) (string, int, error) {
	rctx, ok := ctx.(*tenant.Context)
	if !ok {
		return "", int(view.CacheSkip), nil
	}

	data := map[string]any{
		"IP":      rctx.Request.RemoteAddr,
		"Browser": rctx.UA.Browser,
		"Device":  rctx.UA.Device,
		"OS":      rctx.UA.OS,
		"OSVer":   rctx.UA.OSVersion,
		"IsBot":   rctx.UA.IsBot,
	}

	html, policy, err := view.RenderToString(rctx, "example", "widgets/example", data)
	return string(html), int(policy), err
}

func init() { widget.Register(&Widget{}) }
