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
		baseUrl + "/app": {
			h.Application,
			[]string{"POST", "DELETE", "PUT", "GET"},
		},
		baseUrl + "/app/all": {
			h.Tasks,
			[]string{"GET"},
		},
	}
}
