package v1

import (
	"encoding/json"
	"net/http"
)

// v1 API statuses.
const (
	ACCEPTED = "Accepted"
	LAUNCHED = "Launched"
	FAILED   = "Failed"
	RUNNING  = "Running"
	KILLED   = "Killed"
	NOTFOUND = "Not Found"
	QUEUED   = "Queued"
	UPDATE   = "Updated"
)

// v1 API response format.
type Response struct {
	TaskName string `json:"taskname,omitempty"`
	Message  string `json:"message,omitempty"`
	State    string `json:"state,omitempty"`
}

var (
	InternalServerError func(http.ResponseWriter, Response) = responseFactory(http.StatusInternalServerError)
	BadRequest          func(http.ResponseWriter, Response) = responseFactory(http.StatusBadRequest)
	Success             func(http.ResponseWriter, Response) = responseFactory(http.StatusOK)
)

// Creates a response function.
func responseFactory(statusCode int) func(w http.ResponseWriter, r Response) {
	return func(w http.ResponseWriter, r Response) {
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(r)
	}
}
