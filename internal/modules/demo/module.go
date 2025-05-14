package demo

import (
	"embed"
	htmpl "html/template"

	"github.com/AdeptTravel/adept-framework/internal/requestctx"
	"github.com/AdeptTravel/adept-framework/internal/view"
)

//go:embed views/*.html
var fs embed.FS

// RegisterTemplates adds this module’s templates.
func RegisterTemplates(v *view.Engine) error {
	return v.AppendFS("demo", fs)
}

// Data is now an alias for the unified request context.
type Data = requestctx.RequestCtx

// Render returns the demo fragment.
func Render(v *view.Engine, theme string, d Data) (htmpl.HTML, error) {
	return v.ExecuteToHTML(theme, "demo/demo", d)
}
