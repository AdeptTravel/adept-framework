// internal/form/definition.go
//
// Adept – Forms subsystem: YAML definition loader.
//
// Context
//   Each HTML form in Adept is declared in a YAML file.  This file defines the
//   form’s identifier, title, fields, multi-step structure, and any post-submit
//   actions.  At application start, we parse every “*.yaml” under each
//   “components/<comp>/forms/” directory (plus any tenant overrides) and store
//   the resulting FormDef in an in-memory registry.  Subsequent packages
//   (renderer, validator, actions, widgets) fetch definitions from this
//   registry by ID, guaranteeing a single source of truth.
//
// Workflow
//   •  Structs mirror the YAML schema: FormDef → StepDef → FieldDef / ActionDef.
//   •  LoadFormDef parses a single YAML file and validates structural rules.
//   •  RegisterForms walks one or more base directories, discovers YAMLs,
//      loads them via LoadFormDef, and adds them to the registry, respecting
//      tenant-level override precedence.
//   •  GetFormDef offers safe, read-only access to a parsed form by ID.
//
// Style
//   Comments follow Adept’s guide: full sentences, two spaces after periods,
//   Oxford commas, and clear roles.  Helper comments use short noun phrases.
//
//------------------------------------------------------------------------------

package form

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// -----------------------------------------------------------------------------
// Data structures
// -----------------------------------------------------------------------------

// FormDef represents one form definition loaded from YAML.
//
// The form is uniquely identified by ID which should be namespaced by component,
// e.g. “auth/login”.  A form is defined EITHER by a flat Field list OR by a
// Steps list (multi-step wizard).  Actions are executed after successful
// validation.
type FormDef struct {
	ID      string      `yaml:"id"`      // Component-scoped identifier.
	Title   string      `yaml:"title"`   // Display title, optional.
	Fields  []FieldDef  `yaml:"fields"`  // Flat list of fields (single-step).
	Steps   []StepDef   `yaml:"steps"`   // Multi-step definition.  Mutually exclusive with Fields.
	Actions []ActionDef `yaml:"actions"` // Post-submit actions.  May be empty.
}

// FieldDef describes a single input control on the form.  Validation metadata
// lives inline so the server can enforce the same rules the client hints at.
type FieldDef struct {
	Name        string   `yaml:"name"`        // Submission key.  Required.
	Label       string   `yaml:"label"`       // Human-readable label.  Required.
	Type        string   `yaml:"type"`        // text, email, number, select, checkbox, etc.
	Placeholder string   `yaml:"placeholder"` // Optional placeholder text.
	Required    bool     `yaml:"required"`    // True if input is mandatory.
	MinLength   int      `yaml:"minlength"`   // ≥ 0, 0 means unset.
	MaxLength   int      `yaml:"maxlength"`   // ≥ 0, 0 means unset.
	Pattern     string   `yaml:"pattern"`     // Regex pattern string.
	Options     []string `yaml:"options"`     // For select/radio.  Optional.
	ErrorMsg    string   `yaml:"error"`       // Custom error message, optional.
}

// StepDef groups fields into a wizard step.  At runtime only one step is
// rendered at a time.
type StepDef struct {
	ID     string     `yaml:"id"`    // Unique per form.  If blank, we derive one.
	Title  string     `yaml:"title"` // Display heading, optional.
	Fields []FieldDef `yaml:"fields"`
}

// ActionDef configures an automated action executed after validation.
//
// Action types are loosely typed so new kinds can be introduced without schema
// churn.  Unknown keys are tolerated here; executor code will validate later.
type ActionDef struct {
	Type   string         `yaml:"type"`    // email, store, webhook, pdf, etc.
	Params map[string]any `yaml:",inline"` // Provider-specific fields inline.
}

// -----------------------------------------------------------------------------
// Registry
// -----------------------------------------------------------------------------

// registry maps compositeID (“comp/form”) → *FormDef.  Guarded by mutex.
var (
	registryMu sync.RWMutex
	registry   = make(map[string]*FormDef)
)

// GetFormDef returns a parsed FormDef by composite ID (“component/form”).
// The boolean is false when the ID is unknown.
func GetFormDef(id string) (*FormDef, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	fd, ok := registry[id]
	return fd, ok
}

// -----------------------------------------------------------------------------
// Loader API
// -----------------------------------------------------------------------------

// LoadFormDef parses one YAML file, validates its structure, and returns a
// populated FormDef.  It NEVER mutates the global registry.
func LoadFormDef(path string) (*FormDef, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read form file %s: %w", path, err)
	}

	var fd FormDef
	if err := yaml.Unmarshal(raw, &fd); err != nil {
		return nil, fmt.Errorf("parse YAML %s: %w", path, err)
	}

	if err := validateFormDef(&fd, path); err != nil {
		return nil, err
	}

	return &fd, nil
}

// RegisterForms walks one or more base directories and loads every “*.yaml”
// under “components/*/forms/”.  The dirs slice must be ordered by precedence,
// with tenant override directories BEFORE the global directory.
//
// Example:
//
//	err := form.RegisterForms([]string{
//	    "/var/adept/sites/acme.com", // tenant overrides
//	    "/var/adept",                // global defaults
//	})
func RegisterForms(baseDirs []string) error {
	if len(baseDirs) == 0 {
		return errors.New("RegisterForms: no base directories provided")
	}

	for _, base := range baseDirs {
		formsRoot := filepath.Join(base, "components")
		err := filepath.WalkDir(formsRoot, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() || !strings.HasSuffix(d.Name(), ".yaml") {
				return nil // skip non-YAML
			}

			fd, err := LoadFormDef(path)
			if err != nil {
				return err // fail fast so issues surface loudly.
			}
			register(fd)
			return nil
		})
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err // propagate IO or parse errors.
		}
	}

	return nil
}

// register inserts or overrides the form in the global registry and registers
// a corresponding widget.  Caller must ensure the FormDef passed validation.
func register(fd *FormDef) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[fd.ID] = fd
	injectWidgetRegistration(fd) // ensure widget available for templates.
}

// -----------------------------------------------------------------------------
// Validation helpers
// -----------------------------------------------------------------------------

// validateFormDef enforces structural rules that cannot be expressed via YAML
// tags alone.  It returns a descriptive error referencing the offending file.
func validateFormDef(fd *FormDef, path string) error {
	if fd.ID == "" {
		return fmt.Errorf("form definition %s: missing required 'id'", path)
	}

	// Either flat fields OR steps, not both.
	if len(fd.Fields) > 0 && len(fd.Steps) > 0 {
		return fmt.Errorf("form definition %s: cannot have both 'fields' and 'steps'", path)
	}
	if len(fd.Fields) == 0 && len(fd.Steps) == 0 {
		return fmt.Errorf("form definition %s: must have 'fields' or 'steps'", path)
	}

	// Normalize step IDs if needed and collect field names for uniqueness check.
	fieldNames := make(map[string]struct{})

	if len(fd.Fields) > 0 {
		for i := range fd.Fields {
			if err := validateField(&fd.Fields[i], path); err != nil {
				return err
			}
			if _, dup := fieldNames[fd.Fields[i].Name]; dup {
				return fmt.Errorf("form %s: duplicate field name '%s'", path, fd.Fields[i].Name)
			}
			fieldNames[fd.Fields[i].Name] = struct{}{}
		}
	}

	for si := range fd.Steps {
		s := &fd.Steps[si]
		if s.ID == "" {
			s.ID = fmt.Sprintf("step%d", si+1)
		}
		for fi := range s.Fields {
			if err := validateField(&s.Fields[fi], path); err != nil {
				return err
			}
			if _, dup := fieldNames[s.Fields[fi].Name]; dup {
				return fmt.Errorf("form %s: duplicate field name '%s' across steps", path, s.Fields[fi].Name)
			}
			fieldNames[s.Fields[fi].Name] = struct{}{}
		}
	}

	// Validate actions: ensure known type strings only.  Unknown types are
	// allowed for forward compatibility but emit a warning so developers notice.
	validActions := map[string]bool{
		"email":   true,
		"store":   true,
		"webhook": true,
		"pdf":     true,
	}

	for _, ac := range fd.Actions {
		if !validActions[ac.Type] {
			fmt.Fprintf(os.Stderr, "WARNING: form %s: unrecognized action type '%s'\n", fd.ID, ac.Type)
		}
	}

	return nil
}

// validateField confirms that essential attributes are present and sane.
func validateField(f *FieldDef, path string) error {
	if f.Name == "" {
		return fmt.Errorf("form %s: field missing 'name'", path)
	}
	if f.Label == "" {
		return fmt.Errorf("form %s: field '%s' missing 'label'", path, f.Name)
	}
	if f.Type == "" {
		return fmt.Errorf("form %s: field '%s' missing 'type'", path, f.Name)
	}

	if f.Pattern != "" {
		if _, err := regexp.Compile(f.Pattern); err != nil {
			return fmt.Errorf("form %s: field '%s' invalid regex pattern: %v", path, f.Name, err)
		}
	}

	if f.MinLength < 0 || f.MaxLength < 0 {
		return fmt.Errorf("form %s: field '%s' minlength/maxlength cannot be negative", path, f.Name)
	}
	if f.MaxLength > 0 && f.MinLength > f.MaxLength {
		return fmt.Errorf("form %s: field '%s' minlength greater than maxlength", path, f.Name)
	}

	return nil
}
