// internal/form/validate.go
//
// Adept – Forms subsystem: server-side validation and sanitization.
//
// Context
//   The renderer outputs HTML containing a CSRF token and timestamp.  When the
//   browser posts user input, this file verifies the submission: CSRF, timing,
//   required fields, type constraints, regex patterns, option values, and
//   length limits.  It returns a sanitized map that business logic and actions
//   can trust.
//
// Workflow
//   •  ValidateForm retrieves the FormDef, flattens its FieldDefs, and checks
//      CSRF + render timestamp before per-field validation.
//   •  Each field is validated and sanitized by type.  Errors are captured in
//      []ErrorField so templates can highlight exact issues.
//   •  On success a map[string]any of clean values is returned.
//   •  On failure callers wrap the []ErrorField in validationError (see
//      submit.go) and treat it as a user error, not a 500.
//
// Style
//   Comments follow Adept’s guide: full sentences, two space spacing, Oxford
//   comma, and IDs like “CSRF.”
//
//------------------------------------------------------------------------------

package form

import (
	"fmt"
	"html"
	"mime/multipart"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// -----------------------------------------------------------------------------
// Error types
// -----------------------------------------------------------------------------

// ErrorField describes a single validation failure so the template can render
// a field-level message.
type ErrorField struct {
	Name    string // field name
	Message string // user-facing message
}

// validationError wraps []ErrorField and satisfies the error interface.
//
// It allows callers (HandleSubmit, component handlers) to distinguish user
// input errors from system failures via errors.As / IsValidationError.
type validationError struct{ Fields []ErrorField }

func (ve validationError) Error() string { return "form validation failed" }

// -----------------------------------------------------------------------------
// Public API
// -----------------------------------------------------------------------------

// ValidateForm validates posted form data (already parsed into url.Values) for
// formID.  It returns sanitized values and any field errors.  A non-empty error
// slice means UI re-render is required.
func ValidateForm(formID string, posted url.Values) (map[string]any, []ErrorField) {
	fd, ok := GetFormDef(formID)
	if !ok {
		return nil, []ErrorField{{Name: "", Message: "Unknown form."}}
	}

	var errs []ErrorField
	clean := make(map[string]any)

	// -------------------------------------------------------------------------
	// Form-level checks: CSRF + render timestamp
	// -------------------------------------------------------------------------
	if !verifyCSRF(posted.Get("csrf_token")) {
		errs = append(errs, ErrorField{"", "Security token invalid.  Please refresh and try again."})
		return nil, errs
	}
	if msg := checkTiming(posted.Get("render_ts")); msg != "" {
		errs = append(errs, ErrorField{"", msg})
		return nil, errs
	}

	// -------------------------------------------------------------------------
	// Per-field validation
	// -------------------------------------------------------------------------
	for _, f := range flattenFields(fd) {
		raw, present := extractValue(posted, &f)

		// Required
		if f.Required && !present {
			errs = append(errs, ErrorField{f.Name, requiredMsg(&f)})
			continue
		}
		// Empty optional – nothing more to do.
		if !present || raw == "" {
			continue
		}

		val, perr := validateAndSanitize(&f, raw)
		if perr != "" {
			errs = append(errs, ErrorField{f.Name, perr})
			continue
		}
		clean[f.Name] = val
	}

	return clean, errs
}

// -----------------------------------------------------------------------------
// Form-level helpers
// -----------------------------------------------------------------------------

func verifyCSRF(token string) bool {
	return token != "" && VerifyToken(token)
}

// checkTiming ensures the form was not submitted suspiciously fast or too late.
// Returns empty string on success, user-visible message on failure.
func checkTiming(tsRaw string) string {
	if tsRaw == "" {
		return "Timestamp missing.  Please reload the page."
	}
	ts, err := strconv.ParseInt(tsRaw, 10, 64)
	if err != nil {
		return "Bad timestamp.  Please retry."
	}
	delta := time.Since(time.UnixMicro(ts))
	switch {
	case delta < 2*time.Second:
		return "Form submitted too quickly.  Please enter the fields manually."
	case delta > 30*time.Minute:
		return "Form expired.  Please reload and submit again."
	default:
		return ""
	}
}

// flattenFields returns all FieldDefs regardless of step structure.
func flattenFields(fd *FormDef) []FieldDef {
	if len(fd.Steps) == 0 {
		return fd.Fields
	}
	var out []FieldDef
	for _, s := range fd.Steps {
		out = append(out, s.Fields...)
	}
	return out
}

// -----------------------------------------------------------------------------
// Field-level helpers
// -----------------------------------------------------------------------------

// extractValue obtains the raw submitted value for field f.
func extractValue(v url.Values, f *FieldDef) (string, bool) {
	raw, ok := v[f.Name]
	if !ok || len(raw) == 0 {
		return "", false
	}
	// For checkbox, presence = true; value often "on".
	if f.Type == "checkbox" {
		return raw[0], true
	}
	return raw[0], true
}

func validateAndSanitize(f *FieldDef, raw string) (any, string) {
	val := strings.TrimSpace(raw)

	switch f.Type {
	case "text", "textarea":
		if msg := lengthCheck(f, val); msg != "" {
			return nil, msg
		}
		if f.Pattern != "" && !regexMatch(f.Pattern, val) {
			return nil, patternMsg(f)
		}
		return html.EscapeString(val), ""

	case "email":
		if msg := lengthCheck(f, val); msg != "" {
			return nil, msg
		}
		if _, err := mail.ParseAddress(val); err != nil {
			return nil, invalidMsg(f)
		}
		return val, ""

	case "password":
		if msg := lengthCheck(f, val); msg != "" {
			return nil, msg
		}
		return val, ""

	case "number":
		if msg := lengthCheck(f, val); msg != "" {
			return nil, msg
		}
		if _, err := strconv.ParseFloat(val, 64); err != nil {
			return nil, invalidMsg(f)
		}
		return val, ""

	case "date":
		if _, err := time.Parse("2006-01-02", val); err != nil {
			return nil, invalidMsg(f)
		}
		return val, ""

	case "checkbox":
		// Checked = true, unchecked not present.
		return true, ""

	case "select", "radio":
		if !optionAllowed(f.Options, val) {
			return nil, invalidMsg(f)
		}
		return val, ""

	default:
		return nil, fmt.Sprintf("Unsupported field type %q.", f.Type)
	}
}

// lengthCheck validates minlength / maxlength rules.
func lengthCheck(f *FieldDef, s string) string {
	n := len(s)
	if f.MinLength > 0 && n < f.MinLength {
		return fmt.Sprintf("Must be at least %d characters.", f.MinLength)
	}
	if f.MaxLength > 0 && n > f.MaxLength {
		return fmt.Sprintf("Must be less than %d characters.", f.MaxLength)
	}
	return ""
}

func regexMatch(pattern, s string) bool {
	re, _ := regexp.Compile(pattern) // pattern pre-validated at load
	return re.MatchString(s)
}

func optionAllowed(opts []string, v string) bool {
	for _, o := range opts {
		if o == v {
			return true
		}
	}
	return false
}

// user-friendly default messages
func requiredMsg(f *FieldDef) string {
	if f.ErrorMsg != "" {
		return f.ErrorMsg
	}
	return "This field is required."
}
func invalidMsg(f *FieldDef) string {
	if f.ErrorMsg != "" {
		return f.ErrorMsg
	}
	return "Invalid input."
}
func patternMsg(f *FieldDef) string {
	if f.ErrorMsg != "" {
		return f.ErrorMsg
	}
	return "Input does not match required format."
}

// -----------------------------------------------------------------------------
// Multipart placeholder for future file inputs
// -----------------------------------------------------------------------------

func postedFile(_ *multipart.Reader, _ string) (*multipart.FileHeader, bool) {
	return nil, false // to be implemented with file field support
}
