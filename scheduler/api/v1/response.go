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
	MethodNotAllowed    func(http.ResponseWriter, Response)   = responseFactory(http.StatusMethodNotAllowed)
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
