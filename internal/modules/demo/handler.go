package demo

import (
	"net/http"

	"github.com/AdeptTravel/adept-framework/internal/view"
)

// Handler returns an http.HandlerFunc that renders the demo page.
//   - engine : view engine
//   - dataFn : callback that produces Data for each request
func Handler(engine *view.Engine, dataFn func(*http.Request) Data) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := dataFn(r)
		// theme is hard-coded to "minimal" for now; later you’ll pull from site context.
		if err := engine.Exec(w, "minimal", "demo/demo", data); err != nil {
			http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		}
	}
}
