// internal/form/submit.go
//
// Adept – Forms subsystem: consolidated Submit helper.
//
// Context
//   Most handlers want one call that: parses POST body, validates input,
//   executes configured actions, and returns the clean map or a ValidationError.
//   HandleSubmit provides that convenience so component code stays terse.
//
//------------------------------------------------------------------------------

package form

import (
	"errors"
	"net/http"
)

// HandleSubmit parses r, validates against formID, executes default actions,
// and returns the sanitized data.  On validation failure it returns a
// ValidationError (check with IsValidationError).  On unexpected system
// failures it returns a generic error.
func HandleSubmit(formID string, r *http.Request) (map[string]any, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	clean, errs := ValidateForm(formID, r.PostForm)
	if len(errs) > 0 {
		return nil, validationError{Fields: errs}
	}

	ExecuteActions(formID, clean, ActionCtx{Ctx: r.Context()})
	return clean, nil
}

// IsValidationError reports whether err came from failed ValidateForm.
func IsValidationError(err error) bool {
	var ve validationError
	return errors.As(err, &ve)
}
