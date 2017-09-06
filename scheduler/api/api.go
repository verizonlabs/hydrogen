// Copyright 2017 Verizon
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"mesos-framework-sdk/logging"
	"net/http"
	"os"
	sched "hydrogen/scheduler"
	apiManager "hydrogen/scheduler/api/manager"
	"hydrogen/scheduler/api/v1"
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
			}
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
