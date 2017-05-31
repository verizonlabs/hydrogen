package v1

import (
	"net/http"
)

const baseUrl string = "/v1/api"

func MapRoutes(h *Handlers) map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		baseUrl + "/deploy": h.Deploy,
		baseUrl + "/status": h.State,
		baseUrl + "/tasks":  h.Tasks,
		baseUrl + "/kill":   h.Kill,
		baseUrl + "/update": h.Update,
	}
}
