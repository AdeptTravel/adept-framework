// internal/form/widget.go
//
// Adept – Forms subsystem: widget integration (adjusted signature).
//
// Context
//   Adept templates embed form markup through the widget system.  The concrete
//   widget.Widget interface expects:
//
//       Render(rctx any, params map[string]any) (string, int, error)
//
//   This adapter wraps RenderForm and always returns view.CacheSkip so pages
//   never cache CSRF tokens.
//
//------------------------------------------------------------------------------

package form

import (
	"github.com/yanizio/adept/internal/view"
	"github.com/yanizio/adept/internal/widget"
)

// Ensure compile-time compliance with widget.Widget.
var _ widget.Widget = (*formWidget)(nil)

type formWidget struct{ id string }

// ID implements widget.Widget.
func (w *formWidget) ID() string { return w.id }

// Render converts the FormDef into HTML.  params may include:
//
//   - "prefill" map[string]string – values to pre-populate inputs
//   - "step"    string           – specific step ID in a multi-step form
//
// It always returns view.CacheSkip so every render gets a fresh CSRF token.
func (w *formWidget) Render(_ any, params map[string]any) (string, int, error) {
	var pre map[string]string
	var step string
	if params != nil {
		if p, ok := params["prefill"].(map[string]string); ok {
			pre = p
		}
		if s, ok := params["step"].(string); ok {
			step = s
		}
	}

	htmlOut, err := RenderForm(w.id, RenderOptions{
		Prefill: pre,
		StepID:  step,
	})
	if err != nil {
		return "", int(view.CacheSkip), err
	}
	// template.HTML is an alias of string; cast to satisfy interface.
	return string(htmlOut), int(view.CacheSkip), nil
}

// injectWidgetRegistration is called by definition.go after each FormDef loads.
func injectWidgetRegistration(fd *FormDef) { widget.Register(&formWidget{id: fd.ID}) }
