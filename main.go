package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/pat"
	"github.com/justinas/alice"
	"github.com/justinas/nosurf"
	"github.com/zxchris/swaggerly/config"
	"github.com/zxchris/swaggerly/handlers/guides"
	"github.com/zxchris/swaggerly/handlers/home"
	"github.com/zxchris/swaggerly/handlers/reference"
	"github.com/zxchris/swaggerly/handlers/specs"
	"github.com/zxchris/swaggerly/handlers/static"
	"github.com/zxchris/swaggerly/handlers/timeout"
	"github.com/zxchris/swaggerly/logger"
	"github.com/zxchris/swaggerly/network"
	"github.com/zxchris/swaggerly/proxy"
	"github.com/zxchris/swaggerly/render"
	"github.com/zxchris/swaggerly/spec"
)

var VERSION string

// ---------------------------------------------------------------------------
func main() {

	VERSION = "1.0.0" // TODO build with doxc to control version number?

	log.Printf("Swaggerly server version %s starting\n", VERSION)

	cfg, err := config.Get()
	if err != nil {
		log.Fatalf("error configuring app: %s", err)
	}

	// logging before this point must rely on setting LOGLEVEL env var
	if l, err := logger.LevelFromString(cfg.LogLevel); err == nil {
		logger.DefaultLevel = l
	} else {
		logger.Errorf(nil, "error setting log level: %s", err)
		os.Exit(1)
	}

	router := pat.New()
	chain := alice.New(logger.Handler /*, context.ClearHandler*/, timeoutHandler, withCsrf).Then(router)

	logger.Infof(nil, "listening on %s", cfg.BindAddr)
	listener, err := net.Listen("tcp", cfg.BindAddr)
	if err != nil {
		logger.Errorf(nil, "%s", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	var sg sync.WaitGroup
	sg.Add(1)

	go func() {
		logger.Traceln(nil, "Listen for and serve swagger spec requests for start up")
		wg.Add(1)
		sg.Done()
		http.Serve(listener, chain)
		logger.Traceln(nil, "Finished service swagger specs for start up")
		wg.Done()
	}()

	sg.Wait()

	// Register the spec routes (Listener and server must be up and running by now)
	specs.Register(router)
	spec.LoadStatusCodes()

	err = spec.LoadSpecifications(cfg.BindAddr, true)
	if err != nil {
		logger.Errorf(nil, "%s", err)
		os.Exit(1)
	}

	render.Register()

	reference.Register(router)
	guides.Register(router)
	static.Register(router) // TODO - Static content should be capable of being CDN hosted

	home.Register(router)
	proxy.Register(router)

	listener.Close() // Stop serving specs
	wg.Wait()        // wait for go routine serving specs to terminate

	listener, err = network.GetListener()
	if err != nil {
		logger.Errorf(nil, "Error listening on %s: %s", cfg.BindAddr, err)
		os.Exit(1)
	}

	http.Serve(listener, chain)
}

// ---------------------------------------------------------------------------
func withCsrf(h http.Handler) http.Handler {
	csrfHandler := nosurf.New(h)
	csrfHandler.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rsn := nosurf.Reason(req).Error()
		logger.Warnf(req, "failed csrf validation: %s", rsn)
		render.HTML(w, http.StatusBadRequest, "error", map[string]interface{}{"error": rsn})
	}))
	return csrfHandler
}

// ---------------------------------------------------------------------------
func timeoutHandler(h http.Handler) http.Handler {
	return timeout.Handler(h, 1*time.Second, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		logger.Warnln(req, "request timed out")
		render.HTML(w, http.StatusRequestTimeout, "error", map[string]interface{}{"error": "Request timed out"})
	}))
}

// ---------------------------------------------------------------------------

//type myHandler struct {
//}
//
//func (a *myHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
//	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
//}
//
//func addHeader(h http.Handler) http.Handler {
//	return h
//}

// ---------------------------------------------------------------------------
