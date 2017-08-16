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
