package api

import (
	"encoding/json"
	"mesos-framework-sdk/logging"
	"net/http"
	"os"
	sched "sprint/scheduler"
	apiManager "sprint/scheduler/api/manager"
	"sprint/scheduler/api/v1"
)

// API server provides an interface for users to interact with the core scheduler.
type ApiServer struct {
	cfg     *sched.Configuration
	manager apiManager.ApiParser
	logger  logging.Logger
}

// Returns a new API server injected with the necessary components.
func NewApiServer(cfg *sched.Configuration, mgr apiManager.ApiParser, lgr logging.Logger) *ApiServer {
	return &ApiServer{
		cfg:     cfg,
		manager: mgr,
		logger:  lgr,
	}
}

// Registers an HTTP handler to a given path.
// Applies middleware to determine if the supplied HTTP method is allowed or not.
func (a *ApiServer) applyRoute(path string, route v1.Route) {
	mux := a.cfg.APIServer.Server.Mux()

	// Apply middleware to determine if the HTTP method is allowed or not for each endpoint.
	mux.HandleFunc(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, method := range route.Methods {
			if method == r.Method {
				route.Handler(w, r)
				return
			}

			json.NewEncoder(w).Encode(v1.Response{
				Status:  v1.FAILED,
				Message: r.Method + " is not allowed on this endpoint.",
			})
		}
	}))
}

// Detects the API version to be used and registers the handlers to the server.
func (a *ApiServer) applyRoutes(version string) {
	switch version {
	case "v1":
		routes := v1.MapRoutes(v1.NewHandlers(a.manager))
		for path, route := range routes {
			a.applyRoute(path, route)
		}
	}
}

// RunAPI runs the server, optionally using TLS.
func (a *ApiServer) RunAPI(handlers map[string]http.HandlerFunc) {
	a.applyRoutes(a.cfg.APIServer.Version)
	apiSrvCfg := a.cfg.APIServer.Server

	if apiSrvCfg.TLS() {
		if err := apiSrvCfg.Server().ListenAndServeTLS(apiSrvCfg.Cert(), apiSrvCfg.Key()); err != nil {
			a.logger.Emit(logging.ERROR, err.Error())
			os.Exit(7)
		}
	} else {
		if err := apiSrvCfg.Server().ListenAndServe(); err != nil {
			a.logger.Emit(logging.ERROR, err.Error())
			os.Exit(7)
		}
	}
}
