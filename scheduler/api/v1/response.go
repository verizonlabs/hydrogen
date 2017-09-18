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

// Basic response which the V1 API uses to pass messages back to the user.
type MessageResponse struct {
	Message string `json:"message"`
}

var (
	InternalServerError func(http.ResponseWriter, interface{}) = responseFactory(http.StatusInternalServerError)
	BadRequest          func(http.ResponseWriter, interface{}) = responseFactory(http.StatusBadRequest)
	MethodNotAllowed    func(http.ResponseWriter, interface{}) = responseFactory(http.StatusMethodNotAllowed)
	Success             func(http.ResponseWriter, interface{}) = responseFactory(http.StatusOK)
)

// Creates a response function.
func responseFactory(statusCode int) func(w http.ResponseWriter, r interface{}) {
	return func(w http.ResponseWriter, r interface{}) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(r)
	}
}
