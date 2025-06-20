// internal/form/renderer.go
//
// Adept – Forms subsystem: HTML renderer.
//
// Context
//   Given a parsed FormDef (from definition.go) this package converts the
//   definition into safe, accessible HTML markup.  The renderer supports both
//   single-step and multi-step forms, applies HTML5 validation attributes,
//   injects a CSRF token and render-timestamp hidden inputs, and honours
//   optional pre-fill data (e.g. previously entered values).
//
// Workflow
//   •  RenderForm looks up the FormDef by ID, selects the requested step (if
//      multi-step), and writes each field via writeField.
//   •  Required, minlength, maxlength, pattern, and placeholder attributes are
//      attached where relevant.  Select/radio options are rendered from the
//      YAML Options slice.
//   •  A cryptographically strong CSRF token is generated via csrf.GenerateToken
//      (implemented in csrf.go) and embedded as a hidden <input>.
//   •  A render timestamp is written as microseconds since Unix epoch to allow
//      timing checks during submission validation.
//   •  The caller receives the final HTML as template.HTML so the surrounding
//      template does not double-escape the markup.
//
// Style
//   Output HTML is deliberately plain – no framework classes – so themes can
//   style via element selectors or class hooks.  Each input gets id="fld-{name}"
//   and is wrapped in <div class="form-field"> for consistent styling.
//
//------------------------------------------------------------------------------

package form

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"strconv"
	"strings"
	"time"
)

// RenderOptions bundles optional parameters influencing HTML output.
type RenderOptions struct {
	// Prefill provides initial field values keyed by field name.
	Prefill map[string]string
	// StepID indicates which step of a multi-step form to render.
	// Empty string defaults to first step.
	StepID string
}

// RenderForm returns the HTML markup for the specified form ID.  It selects the
// correct step (or the flat field list) and embeds security tokens.
// Callers typically pass the resulting template.HTML into a widget response.
func RenderForm(formID string, opts RenderOptions) (template.HTML, error) {
	fd, ok := GetFormDef(formID)
	if !ok {
		return "", fmt.Errorf("RenderForm: unknown form %q", formID)
	}

	// Resolve which fields to render based on step selection.
	fields, stepIndex, err := selectFields(fd, opts.StepID)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	// Form wrapper div to allow per-form CSS targeting if desired.
	buf.WriteString(`<div class="adept-form">` + "\n")

	// Iterate fields in definition order.
	for _, f := range fields {
		if err := writeField(&buf, &f, opts.Prefill); err != nil {
			return "", err
		}
	}

	// Hidden meta inputs.
	csrfToken := csrfGenerateToken() // implemented in csrf.go
	buf.WriteString(fmt.Sprintf(`<input type="hidden" name="csrf_token" value="%s">`+"\n", csrfToken))
	buf.WriteString(fmt.Sprintf(`<input type="hidden" name="render_ts" value="%d">`+"\n", time.Now().UnixMicro()))
	if stepIndex >= 0 {
		buf.WriteString(fmt.Sprintf(`<input type="hidden" name="current_step" value="%s">`+"\n", html.EscapeString(fd.Steps[stepIndex].ID)))
	}

	buf.WriteString(`</div>`)
	return template.HTML(buf.String()), nil
}

// selectFields returns the slice of FieldDefs to render for the requested step.
// If the form is single-step, the entire Fields list is returned.  For
// multi-step forms, the step is chosen by ID (or first if StepID is blank).
func selectFields(fd *FormDef, stepID string) ([]FieldDef, int, error) {
	if len(fd.Steps) == 0 {
		return fd.Fields, -1, nil // single-step
	}

	if stepID == "" {
		return fd.Steps[0].Fields, 0, nil
	}
	for i, s := range fd.Steps {
		if s.ID == stepID {
			return s.Fields, i, nil
		}
	}
	return nil, -1, fmt.Errorf("RenderForm: step %q not found in form %q", stepID, fd.ID)
}

// writeField emits HTML for an individual field into buf, applying prefill and
// validation attributes.  Each field is wrapped in a <div class="form-field">.
func writeField(buf *bytes.Buffer, f *FieldDef, prefill map[string]string) error {
	val := prefillValue(f.Name, prefill)

	// Container
	buf.WriteString(`<div class="form-field">` + "\n")

	// Shared attributes
	idAttr := `id="fld-` + html.EscapeString(f.Name) + `"`
	nameAttr := `name="` + html.EscapeString(f.Name) + `"`

	// Label first (for accessibility)
	buf.WriteString(`<label for="fld-` + html.EscapeString(f.Name) + `">` + html.EscapeString(f.Label) + `</label>` + "\n")

	switch f.Type {
	case "text", "email", "password", "number", "date":
		buf.WriteString(`<input ` + idAttr + ` ` + nameAttr + ` type="` + f.Type + `"`)
		if f.Placeholder != "" {
			buf.WriteString(` placeholder="` + html.EscapeString(f.Placeholder) + `"`)
		}
		if f.Required {
			buf.WriteString(` required`)
		}
		if f.MinLength > 0 {
			buf.WriteString(` minlength="` + strconv.Itoa(f.MinLength) + `"`)
		}
		if f.MaxLength > 0 {
			buf.WriteString(` maxlength="` + strconv.Itoa(f.MaxLength) + `"`)
		}
		if f.Pattern != "" {
			buf.WriteString(` pattern="` + html.EscapeString(f.Pattern) + `"`)
		}
		if val != "" {
			// password fields are not prefilled.
			if f.Type != "password" {
				buf.WriteString(` value="` + html.EscapeString(val) + `"`)
			}
		}
		buf.WriteString(`>` + "\n")

	case "textarea":
		buf.WriteString(`<textarea ` + idAttr + ` ` + nameAttr)
		if f.Required {
			buf.WriteString(` required`)
		}
		if f.MinLength > 0 {
			buf.WriteString(` minlength="` + strconv.Itoa(f.MinLength) + `"`)
		}
		if f.MaxLength > 0 {
			buf.WriteString(` maxlength="` + strconv.Itoa(f.MaxLength) + `"`)
		}
		if f.Placeholder != "" {
			buf.WriteString(` placeholder="` + html.EscapeString(f.Placeholder) + `"`)
		}
		buf.WriteString(`>`)
		if val != "" {
			buf.WriteString(html.EscapeString(val))
		}
		buf.WriteString(`</textarea>` + "\n")

	case "select":
		buf.WriteString(`<select ` + idAttr + ` ` + nameAttr)
		if f.Required {
			buf.WriteString(` required`)
		}
		buf.WriteString(`>` + "\n")
		for _, opt := range f.Options {
			sel := ""
			if val == opt {
				sel = ` selected`
			}
			buf.WriteString(`<option value="` + html.EscapeString(opt) + `"` + sel + `>` + html.EscapeString(opt) + `</option>` + "\n")
		}
		buf.WriteString(`</select>` + "\n")

	case "checkbox":
		checked := ""
		if val != "" && strings.ToLower(val) != "false" {
			checked = ` checked`
		}
		buf.WriteString(`<input ` + idAttr + ` ` + nameAttr + ` type="checkbox"` + checked)
		if f.Required {
			buf.WriteString(` required`)
		}
		buf.WriteString(`>` + "\n")

	case "radio":
		// Render each option as separate radio input
		for i, opt := range f.Options {
			radioID := fmt.Sprintf("fld-%s-%d", f.Name, i)
			checked := ""
			if val == opt {
				checked = ` checked`
			}
			buf.WriteString(`<div class="radio-option">` + "\n")
			buf.WriteString(`<input id="` + radioID + `" name="` + html.EscapeString(f.Name) + `" type="radio" value="` + html.EscapeString(opt) + `"` + checked)
			if f.Required {
				buf.WriteString(` required`)
			}
			buf.WriteString(`>` + "\n")
			buf.WriteString(`<label for="` + radioID + `">` + html.EscapeString(opt) + `</label>` + "\n")
			buf.WriteString(`</div>` + "\n")
		}

	default:
		return fmt.Errorf("writeField: unsupported field type %q in form field %s", f.Type, f.Name)
	}

	// Placeholder span for error messages (populated client-side or server re-render).
	buf.WriteString(`<span class="error" aria-live="polite"></span>` + "\n")

	buf.WriteString(`</div>` + "\n")
	return nil
}

// prefillValue returns previously submitted value or empty string.
func prefillValue(name string, pre map[string]string) string {
	if pre == nil {
		return ""
	}
	return pre[name]
}

// csrfGenerateToken is a thin wrapper around the CSRF package so renderer does
// not import it.  This avoids an import cycle (renderer → csrf → renderer).
func csrfGenerateToken() string {
	token, err := GenerateToken() // from csrf.go
	if err != nil {
		// Fall back to timestamp-based token on unexpected failure (extremely rare).
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return token
}
