// components/auth/auth.go
//
// Adept authentication component – login flow.
//
//------------------------------------------------------------------------------

package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/yanizio/adept/internal/component"
	"github.com/yanizio/adept/internal/form"
	"github.com/yanizio/adept/internal/session"
	"github.com/yanizio/adept/internal/tenant"
	"github.com/yanizio/adept/internal/view"
)

// Compile-time assertion: *Component satisfies component.Component.
var _ component.Component = (*Component)(nil)

// Component encapsulates login functionality.
type Component struct{}

/*────────────────── component.Component methods ───────────────────────────*/

// Name returns the canonical component key.
func (c *Component) Name() string { return "auth" }

// Migrations returns nil – auth has no DB schema (yet).
func (c *Component) Migrations() []string { return nil }

// Init satisfies component.Initializer; no tenant-specific boot work.
func (c *Component) Init(component.TenantInfo) error { return nil }

// Routes builds and returns the router mounted at “/”.
func (c *Component) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/login", c.handleLoginGET)
	r.Post("/login", c.handleLoginPOST)
	return r
}

// Register component at program start.
func init() { component.Register(&Component{}) }

/*──────────────────────────── Handlers ─────────────────────────────────────*/

func (c *Component) handleLoginGET(w http.ResponseWriter, r *http.Request) {
	vctx := tenant.NewContext(r) // *tenant.Context for view layer
	view.Render(vctx, w, "auth", "login.html", nil, view.CacheSkip)
}

func (c *Component) handleLoginPOST(w http.ResponseWriter, r *http.Request) {
	vctx := tenant.NewContext(r)

	data, err := form.HandleSubmit("auth/login", r)
	if err != nil {
		if form.IsValidationError(err) {
			view.Render(vctx, w, "auth", "login.html", map[string]any{
				"FormErrors":  err,
				"FormPrefill": r.PostForm,
			}, view.CacheSkip)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	email := data["email"].(string)
	pass := data["password"].(string)
	if !checkCredentials(email, pass) {
		view.Render(vctx, w, "auth", "login.html", map[string]any{
			"FormErrors": []form.ErrorField{{
				Name:    "password",
				Message: "Incorrect email or password.",
			}},
			"FormPrefill": r.PostForm,
		}, view.CacheSkip)
		return
	}

	session.LoginUser(w, r, email)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

/*──────────────────────── Stub credential check ───────────────────────────*/

func checkCredentials(_, _ string) bool { return false }
