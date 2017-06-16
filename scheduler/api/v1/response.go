package v1

import (
	"encoding/json"
	"net/http"
)

// v1 API response format.
type Response struct {
	TaskName string `json:"taskname,omitempty"`
	Message  string `json:"message,omitempty"`
	State    string `json:"state,omitempty"`
}

var (
	InternalServerError func(http.ResponseWriter, Response)   = responseFactory(http.StatusInternalServerError)
	BadRequest          func(http.ResponseWriter, Response)   = responseFactory(http.StatusBadRequest)
	MethodNotAllowed    func(http.ResponseWriter, Response) = responseFactory(http.StatusMethodNotAllowed)
	Success             func(http.ResponseWriter, Response)   = responseFactory(http.StatusOK)
	MultiSuccess        func(http.ResponseWriter, []Response) = multiResponseFactory(http.StatusOK)
)

// Creates a response function.
func responseFactory(statusCode int) func(w http.ResponseWriter, r Response) {
	return func(w http.ResponseWriter, r Response) {
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(r)
	}
}

// Creates a response function that handles multiple responses.
func multiResponseFactory(statusCode int) func(w http.ResponseWriter, r []Response) {
	return func(w http.ResponseWriter, r []Response) {
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(r)
	}
}
