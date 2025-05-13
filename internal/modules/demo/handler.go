package demo

import (
	"net/http"

	"github.com/AdeptTravel/adept-framework/internal/view"
)

func Handler(engine *view.Engine, getData func(*http.Request) Data) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		d := getData(r)                       // extract timestamp, IP, geo …
		if err := engine.Exec(w, "minimal", "reqinfo/debug", d); err != nil {
			http.Error(w, "template error: "+err.Error(), 500)
		}
	}
}
