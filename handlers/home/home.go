package home

import (
	"net/http"

	"github.com/gorilla/pat"
	"github.com/zxchris/swaggerly/config"
	"github.com/zxchris/swaggerly/logger"
	"github.com/zxchris/swaggerly/render"
	"github.com/zxchris/swaggerly/spec"
)

// ----------------------------------------------------------------------------------------
// Register creates routes for each home handler
func Register(r *pat.Router) {
	logger.Debugln(nil, "registering handlers for home page")

	count := 0
	// Homepages for each loaded specification
	for _, specification := range spec.APISuite {

		logger.Tracef(nil, "Build homepage route for specification '%s'", specification.ID)

		r.Path("/" + specification.ID + "/").Methods("GET").HandlerFunc(specHomeHandler(specification))

		count++
	}

	cfg, _ := config.Get()

	if count == 1 && cfg.ForceRootPage == false {
		// If there is only one specification loaded, then hotwire '/' to redirect to that index page
		// unless Swaggerly is configured not to do that!
		r.Path("/").Methods("GET").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			http.Redirect(w, req, "/uber-api/", 302)
		})
	} else {
		r.Path("/").Methods("GET").HandlerFunc(topHandler) // Top level homepage
	}
}

// ----------------------------------------------------------------------------------------
// Handler is a http.Handler for the home page
func topHandler(w http.ResponseWriter, req *http.Request) {
	logger.Printf(nil, "Render HTML for top level index page")

	render.HTML(w, http.StatusOK, "index", render.DefaultVars(req, nil, render.Vars{"Title": "API documentation"}))
}

// ----------------------------------------------------------------------------------------
func specHomeHandler(specification *spec.APISpecification) func(w http.ResponseWriter, req *http.Request) {

	// The default "theme" level index page.
	tmpl := "specindex"

	//customTmpl := specification.ID + "/" + tmpl
	customTmpl := specification.ID + "/index" // The spec-level indexes are 'index.tmpl'

	logger.Tracef(nil, "+ Test for template '%s'", customTmpl)

	if render.TemplateLookup(customTmpl) != nil {
		tmpl = customTmpl
	}
	return func(w http.ResponseWriter, req *http.Request) {
		render.HTML(w, http.StatusOK, tmpl, render.DefaultVars(req, specification, render.Vars{"Title": "API documentation"}))
	}
}

// ----------------------------------------------------------------------------------------
// end
