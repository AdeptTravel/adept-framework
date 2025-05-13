package demo

import (
	"embed"
	htmpl "html/template"

	"github.com/AdeptTravel/adept-framework/internal/site"
	"github.com/AdeptTravel/adept-framework/internal/view"
)

//go:embed views/*.html
var fs embed.FS

// RegisterTemplates adds this module’s templates under the "demo/" namespace.
func RegisterTemplates(v *view.Engine) error {
	return v.AppendFS("demo", fs)
}

// Data is just an alias of site.ReqInfo, so callers can pass it directly.
type Data = site.ReqInfo

// Render turns Data into an HTML fragment (useful for widgets later).
func Render(v *view.Engine, theme string, d Data) (htmpl.HTML, error) {
	return v.ExecuteToHTML(theme, "views/demo", d)
}
