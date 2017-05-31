package v1

import (
	"net/http"
)

const baseUrl string = "/v1/api"

type Route struct {
	Handler http.HandlerFunc
	Methods []string
}

// Returns a mapping of routes to their respective handlers.
func MapRoutes(h *Handlers) map[string]Route {
	return map[string]Route{
		baseUrl + "/deploy": {
			h.Deploy,
			[]string{"POST"},
		},
		baseUrl + "/status": {
			h.State,
			[]string{"GET"},
		},
		baseUrl + "/tasks": {
			h.Tasks,
			[]string{"GET"},
		},
		baseUrl + "/kill": {
			h.Kill,
			[]string{"DELETE"},
		},
		baseUrl + "/update": {
			h.Update,
			[]string{"PUT"},
		},
	}
}
