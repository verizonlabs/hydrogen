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
	"errors"
	mockLogger "mesos-framework-sdk/logging/test"
	"github.com/verizonlabs/hydrogen/scheduler"
	mockApiManager "github.com/verizonlabs/hydrogen/scheduler/api/manager/test"
	"testing"
)

type brokenReader struct{}

func (b brokenReader) Read(n []byte) (int, error) {
	return 0, errors.New("I'm broke.")
}

var (
	c      = new(scheduler.Configuration)
	l      = new(mockLogger.MockLogger)
	apiMgr = new(mockApiManager.MockApiManager)
)

// Ensures all components are set correctly when creating the API server.
func TestNewApiServer(t *testing.T) {
	srv := NewApiServer(c, apiMgr, l)
	if srv.cfg != c || srv.manager != apiMgr || srv.logger != l {
		t.Fatal("API does not contain the correct components")
	}
}
