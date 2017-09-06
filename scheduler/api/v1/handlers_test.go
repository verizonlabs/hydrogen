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
	"io"
	"mesos-framework-sdk/resources/manager/test"
	test3 "mesos-framework-sdk/scheduler/test"
	"net/http"
	"net/http/httptest"
	"hydrogen/scheduler/api/manager"
	mockApiManager "hydrogen/scheduler/api/manager/test"
	test2 "hydrogen/task/manager/test"
	"strings"
	"testing"
)

var (
	apiMgr                = new(mockApiManager.MockApiManager)
	brokenApiMgr          = new(mockApiManager.MockBrokenApiManager)
	validJSON             = `[{"name": "test", "resources": {"cpu": 0.5, "mem": 128.0}, "command": {"cmd": "echo hello"}}]`
	killJSON              = `{"name": "test"}`
	junkJSON              = `not even json, how did this even get here`
	filtersJSON           = `{"name": "test", "filters": [{"type": "TEXT", "value": ["tester"]}], "resources": {"cpu": 0.5, "mem": 128.0}, "command": {"cmd": "echo hello"}}`
	badFiltersJSON        = `{"name": "test", "filters": [{"type": "not real", "value": "tester"}], "resources": {"cpu": 0.5, "mem": 128.0}, "command": {"cmd": "echo hello"}}`
	invalidFilterTypeJSON = `{"name": "test", "filters": [{"type": "fake news", "value": ["tester"]}], "resources": {"cpu": 0.5, "mem": 128.0}, "command": {"cmd": "echo hello"}}`
)

// Common request-response handling for tests to share.
func requestFixture(f http.HandlerFunc, method, endpoint string, r io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, endpoint, r)

	rr := httptest.NewRecorder()
	http.HandlerFunc(f).ServeHTTP(rr, req)

	return rr
}

// Verifies that the handlers have the correct state.
func TestNewHandlers(t *testing.T) {
	h := NewHandlers(apiMgr)
	if h.manager != apiMgr {
		t.Fatal("API does not contain the correct components")
	}
}

// Validates the deployment endpoint.
func TestHandlers_Deploy(t *testing.T) {
	h := NewHandlers(apiMgr)
	h.manager = mockApiManager.MockApiManager{}
	rr := requestFixture(h.Application, "POST", "/app", strings.NewReader(validJSON))
	if rr.Code != http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", http.StatusOK, rr.Code)
	}
}

// Makes sure the deployment endpoint gives an error when it should.
func TestHandlers_DeployError(t *testing.T) {
	h := NewHandlers(brokenApiMgr)
	h.manager = manager.NewApiParser(
		&test.MockResourceManager{},
		&test2.MockTaskManager{},
		test3.MockScheduler{},
	)
	rr := requestFixture(h.Application, "POST", "/app", strings.NewReader(junkJSON))
	if rr.Code == http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", rr.Code, http.StatusOK)
	}
}

// Validates the endpoint to kill tasks.
func TestHandlers_Kill(t *testing.T) {
	h := NewHandlers(apiMgr)
	h.manager = mockApiManager.MockApiManager{}
	rr := requestFixture(h.Application, "DELETE", "/app", strings.NewReader(killJSON))
	if rr.Code != http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", http.StatusOK, rr.Code)
	}
}

// Makes sure the endpoint to kill tasks gives an error when it should.
func TestHandlers_KillError(t *testing.T) {
	h := NewHandlers(brokenApiMgr)
	h.manager = mockApiManager.MockBrokenApiManager{}
	rr := requestFixture(h.Application, "DELETE", "/app", strings.NewReader(killJSON))
	if rr.Code == http.StatusOK {
		t.Fatalf("Wrong status code: got %d instead of 500", http.StatusOK)
	}
}

// Validates the endpoint to get task state.
func TestHandlers_State(t *testing.T) {
	h := NewHandlers(apiMgr)
	h.manager = mockApiManager.MockApiManager{}
	rr := requestFixture(h.Application, "GET", "/app?name=test", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", http.StatusOK, rr.Code)
	}
}

// Makes sure the endpoint to get task state gives an error when it should.
func TestHandlers_StateError(t *testing.T) {
	h := NewHandlers(brokenApiMgr)
	h.manager = mockApiManager.MockBrokenApiManager{}
	rr := requestFixture(h.Application, "GET", "/app?name=test", nil)
	if rr.Code == http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", rr.Code, http.StatusOK)
	}

	h = NewHandlers(apiMgr)
	h.manager = mockApiManager.MockApiManager{}
	rr = requestFixture(h.Application, "GET", "/app", nil)
	if rr.Code == http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", rr.Code, http.StatusOK)
	}
}

// Validates the endpoint to get all tasks.
func TestHandlers_Tasks(t *testing.T) {
	h := NewHandlers(apiMgr)
	h.manager = mockApiManager.MockApiManager{}
	rr := requestFixture(h.Tasks, "GET", "/app/all", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", http.StatusOK, rr.Code)
	}
}

// Tests that we get an OK response to an empty task manager.
func TestHandlers_TasksEmpty(t *testing.T) {
	h := NewHandlers(brokenApiMgr)
	h.manager = mockApiManager.MockApiManager{}
	rr := requestFixture(h.Tasks, "GET", "/app/all", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", rr.Code, http.StatusOK)
	}
}

// Validates the endpoint to update a task.
func TestHandlers_Update(t *testing.T) {
	h := NewHandlers(apiMgr)
	h.manager = mockApiManager.MockApiManager{}
	rr := requestFixture(h.Application, "PUT", "/app", strings.NewReader(validJSON))
	if rr.Code != http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", http.StatusOK, rr.Code)
	}
}

// Makes sure our endpoint to update a task gives an error when it should.
func TestHandlers_UpdateError(t *testing.T) {
	h := NewHandlers(brokenApiMgr)
	h.manager = mockApiManager.MockBrokenApiManager{}
	rr := requestFixture(h.Application, "PUT", "/app", strings.NewReader(junkJSON))
	if rr.Code == http.StatusOK {
		t.Fatalf("Wrong status code: want %d but got %d", 400, http.StatusOK)
	}
}
