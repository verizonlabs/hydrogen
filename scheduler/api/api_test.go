package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mesos-framework-sdk/include/mesos"
	sched "mesos-framework-sdk/include/scheduler"
	"mesos-framework-sdk/structures"
	"mesos-framework-sdk/task"
	"net/http"
	"net/http/httptest"
	"sprint/scheduler/api/response"
	"strings"
	"testing"
)

// TODO think about how/where to keep these mocks as other tests start using them
// It sounds like a common pattern is to make a place for mocks and put them all in their own package(s).
// Since you cannot reference types in tests from tests in other packages people seem to be putting them in non-test files.

type mockScheduler struct{}

func (m *mockScheduler) Subscribe(chan *sched.Event) (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Teardown() (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Accept(offerIds []*mesos_v1.OfferID, tasks []*mesos_v1.Offer_Operation, filters *mesos_v1.Filters) (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Decline(offerIds []*mesos_v1.OfferID, filters *mesos_v1.Filters) (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Revive() (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Kill(taskId *mesos_v1.TaskID, agentid *mesos_v1.AgentID) (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Shutdown(execId *mesos_v1.ExecutorID, agentId *mesos_v1.AgentID) (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Acknowledge(agentId *mesos_v1.AgentID, taskId *mesos_v1.TaskID, uuid []byte) (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Reconcile(tasks []*mesos_v1.TaskInfo) (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Message(agentId *mesos_v1.AgentID, executorId *mesos_v1.ExecutorID, data []byte) (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) SchedRequest(resources []*mesos_v1.Request) (*http.Response, error) {
	return new(http.Response), nil
}

func (m *mockScheduler) Suppress() (*http.Response, error) {
	return new(http.Response), nil
}

type mockTaskManager struct{}

func (m *mockTaskManager) Add(*mesos_v1.TaskInfo) error {
	return nil
}

func (m *mockTaskManager) Delete(*mesos_v1.TaskInfo) {

}

func (m *mockTaskManager) Get(*string) (*mesos_v1.TaskInfo, error) {
	return &mesos_v1.TaskInfo{}, nil
}

func (m *mockTaskManager) GetById(id *mesos_v1.TaskID) (*mesos_v1.TaskInfo, error) {
	return &mesos_v1.TaskInfo{}, nil
}

func (m *mockTaskManager) HasTask(*mesos_v1.TaskInfo) bool {
	return false
}

func (m *mockTaskManager) Set(mesos_v1.TaskState, *mesos_v1.TaskInfo) {

}

func (m *mockTaskManager) GetState(state mesos_v1.TaskState) ([]*mesos_v1.TaskInfo, error) {
	return []*mesos_v1.TaskInfo{
		{},
	}, nil
}

func (m *mockTaskManager) TotalTasks() int {
	return 0
}

func (m *mockTaskManager) Tasks() *structures.ConcurrentMap {
	return structures.NewConcurrentMap()
}

type mockResourceManager struct{}

func (m *mockResourceManager) AddOffers(offers []*mesos_v1.Offer) {

}

func (m *mockResourceManager) HasResources() bool {
	return false
}

func (m *mockResourceManager) AddFilter(t *mesos_v1.TaskInfo, filters []task.Filter) error {
	return nil
}

func (m *mockResourceManager) ClearFilters(t *mesos_v1.TaskInfo) {

}

func (m *mockResourceManager) Assign(task *mesos_v1.TaskInfo) (*mesos_v1.Offer, error) {
	return &mesos_v1.Offer{}, nil
}

func (m *mockResourceManager) Offers() []*mesos_v1.Offer {
	return []*mesos_v1.Offer{
		{},
	}
}

type mockLogger struct{}

func (m *mockLogger) Emit(severity uint8, template string, args ...interface{}) {

}

type mockServerConfiguration struct{}

func (m *mockServerConfiguration) Cert() string {
	return ""
}

func (m *mockServerConfiguration) Key() string {
	return ""
}

func (m *mockServerConfiguration) Port() int {
	return 9999
}

func (m *mockServerConfiguration) Path() string {
	return ""
}

func (m *mockServerConfiguration) Protocol() string {
	return "http"
}

func (m *mockServerConfiguration) Server() *http.Server {
	return &http.Server{}
}

func (m *mockServerConfiguration) TLS() bool {
	return false
}

var (
	c              = new(mockServerConfiguration)
	s              = new(mockScheduler)
	tm             = new(mockTaskManager)
	r              = new(mockResourceManager)
	h              = http.NewServeMux()
	v              = "test"
	l              = new(mockLogger)
	validJSON      = fmt.Sprint(`{"name": "test", "resources": {"cpu": 0.5, "mem": 128.0}, "command": {"cmd": "echo hello"}}`)
	killJSON       = fmt.Sprint(`{"name": "test"}`)
	junkJSON       = fmt.Sprint(`not even json, how did this even get here`)
	filtersJSON    = fmt.Sprint(`{"name": "test", "filters": [{"type": "TEXT", "value": ["tester"]}], "resources": {"cpu": 0.5, "mem": 128.0}, "command": {"cmd": "echo hello"}}`)
	badFiltersJSON = fmt.Sprint(`{"name": "test", "filters": [{"type": "not real", "value": "tester"}], "resources": {"cpu": 0.5, "mem": 128.0}, "command": {"cmd": "echo hello"}}`)
)

// Ensures all components are set correctly when creating the API server.
func TestNewApiServer(t *testing.T) {
	srv := NewApiServer(c, s, tm, r, h, v, l)
	if srv.cfg != c || srv.sched != s || srv.taskMgr != tm || srv.resourceMgr != r ||
		srv.mux != h || srv.version != v || srv.logger != l {
		t.Fatal("API does not contain the correct components")
	}
}

func TestNewApiServerRun(t *testing.T) {
	srv := NewApiServer(c, s, tm, r, h, v, l)
	if srv.cfg != c || srv.sched != s || srv.taskMgr != tm || srv.resourceMgr != r ||
		srv.mux != h || srv.version != v || srv.logger != l {
		t.Fatal("API does not contain the correct components")
	}
}

// Checks if our internal handlers are attached correctly.
func TestApiServer_Handle(t *testing.T) {
	srv := NewApiServer(c, s, tm, r, h, v, l)
	handles := map[string]http.HandlerFunc{
		"test1": func(w http.ResponseWriter, r *http.Request) {},
		"test2": func(w http.ResponseWriter, r *http.Request) {},
	}
	srv.setHandlers(handles)

	h := srv.Handle()
	if len(h) != len(handles) {
		t.Fatal("Not all handlers were applied correctly")
	}
}

func TestApiDeploy(t *testing.T) {
	srv := NewApiServer(c, s, tm, r, h, v, l)
	srv.setDefaultHandlers()

	a := strings.NewReader(validJSON)
	req := httptest.NewRequest("POST", "http://127.0.0.1:9999/v1/api/deploy", a)
	w := httptest.NewRecorder()

	srv.deploy(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m response.Deploy
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status == response.FAILED {
		t.Logf("Task shouldn't of failed but did %v", m.Message)
		t.FailNow()
	}
}

func TestApiJunkDeploy(t *testing.T) {
	srv := NewApiServer(c, s, tm, r, h, v, l)
	srv.setDefaultHandlers()
	req := httptest.NewRequest("POST", "localhost:9999/v1/api/deploy", strings.NewReader(
		fmt.Sprint(`{"test":"something"}`),
	))
	w := httptest.NewRecorder()
	srv.deploy(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m response.Deploy
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != response.FAILED {
		t.FailNow()
	}
}

func TestApiUpdate(t *testing.T) {
	srv := NewApiServer(c, s, tm, r, h, v, l)
	srv.setDefaultHandlers()

	a := strings.NewReader(validJSON)
	req := httptest.NewRequest("PUT", "http://127.0.0.1:9999/v1/api/update", a)
	w := httptest.NewRecorder()

	srv.update(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m response.Kill
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != response.UPDATE {
		t.Logf("Task should of been UPDATE but didn't %v", m.Message)
		t.FailNow()
	}
}

func TestApiUpdateFail(t *testing.T) {
	srv := NewApiServer(c, s, tm, r, h, v, l)
	srv.setDefaultHandlers()

	a := strings.NewReader(junkJSON)
	req := httptest.NewRequest("PUT", "http://127.0.0.1:9999/v1/api/update", a)
	w := httptest.NewRecorder()

	srv.update(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m response.Kill
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != response.FAILED {
		t.Logf("Task should of been FAILED but wasn't %v", m.Message)
		t.FailNow()
	}
}

func TestApiKill(t *testing.T) {
	srv := NewApiServer(c, s, tm, r, h, v, l)
	srv.setDefaultHandlers()

	a := strings.NewReader(killJSON)
	req := httptest.NewRequest("DELETE", "http://127.0.0.1:9999/v1/api/kill", a)
	w := httptest.NewRecorder()

	srv.kill(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m response.Kill
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != response.KILLED {
		t.Logf("Task shouldn't of failed but did %v", m.Message)
		t.FailNow()
	}
}

func TestApiKillFail(t *testing.T) {
	srv := NewApiServer(c, s, tm, r, h, v, l)
	srv.setDefaultHandlers()

	a := strings.NewReader(junkJSON)
	req := httptest.NewRequest("DELETE", "http://127.0.0.1:9999/v1/api/kill", a)
	w := httptest.NewRecorder()

	srv.kill(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m response.Kill
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != response.FAILED {
		t.Logf("Task should of failed but didn't:  %v", m.Message)
		t.FailNow()
	}
}

func TestApiState(t *testing.T) {
	srv := NewApiServer(c, s, tm, r, h, v, l)
	srv.setDefaultHandlers()

	req := httptest.NewRequest("GET", "http://127.0.0.1:9999/v1/api/status?name=test", nil)
	w := httptest.NewRecorder()

	srv.state(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m response.Deploy
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != response.LAUNCHED {
		t.Logf("Task should of been in state LAUNCHED but wasn't:  %v", m.Message)
		t.FailNow()
	}
}

func TestApiStateFail(t *testing.T) {
	srv := NewApiServer(c, s, tm, r, h, v, l)
	srv.setDefaultHandlers()

	req := httptest.NewRequest("GET", "http://127.0.0.1:9999/v1/api/status?junkvalue", nil)
	w := httptest.NewRecorder()

	srv.state(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m response.Deploy
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != response.FAILED {
		t.Logf("API should of returned FAILED but didn't:  %v", m.Message)
		t.FailNow()
	}
}

func TestApiFailWrongMethod(t *testing.T) {
	srv := NewApiServer(c, s, tm, r, h, v, l)
	srv.setDefaultHandlers()

	req := httptest.NewRequest("GET", "http://127.0.0.1:9999/v1/api/kill", nil)
	w := httptest.NewRecorder()

	srv.kill(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m response.Deploy
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != response.FAILED {
		t.Logf("API should of returned FAILED but didn't:  %v", m.Message)
		t.FailNow()
	}
}

func TestApiDeployWithNil(t *testing.T) {
	srv := NewApiServer(c, s, tm, r, h, v, l)
	srv.setDefaultHandlers()

	req := httptest.NewRequest("POST", "http://127.0.0.1:9999/v1/api/deploy", nil)
	w := httptest.NewRecorder()

	srv.deploy(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m response.Deploy
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != response.FAILED {
		t.Logf("Task shouldn't of failed but did %v", m.Message)
		t.FailNow()
	}
}

func TestApiDeployWithFilters(t *testing.T) {
	srv := NewApiServer(c, s, tm, r, h, v, l)
	srv.setDefaultHandlers()

	a := strings.NewReader(filtersJSON)
	req := httptest.NewRequest("POST", "http://127.0.0.1:9999/v1/api/deploy", a)
	w := httptest.NewRecorder()

	srv.deploy(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m response.Deploy
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status == response.FAILED {
		t.Logf("Task shouldn't of failed but did %v", m.Message)
		t.FailNow()
	}
}

func TestApiDeployWithFiltersFail(t *testing.T) {
	srv := NewApiServer(c, s, tm, r, h, v, l)
	srv.setDefaultHandlers()

	a := strings.NewReader(badFiltersJSON)
	req := httptest.NewRequest("POST", "http://127.0.0.1:9999/v1/api/deploy", a)
	w := httptest.NewRecorder()

	srv.deploy(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m response.Deploy
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != response.FAILED {
		t.Logf("Task should of failed but didn't %v", m.Message)
		t.FailNow()
	}
}

// TODO (tim): Fix how Stats end point works.
// Concurrent map needs to be mocks since we cast a type from it in stats to get it's State value.
/*
func TestApiStats(t *testing.T) {
	srv := NewApiServer(c, s, tm, r, h, v, l)
	srv.setDefaultHandlers()

	req := httptest.NewRequest("GET", "http://127.0.0.1:9999/v1/api/stats?name=test", nil)
	w := httptest.NewRecorder()

	srv.stats(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m response.Deploy
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != response.ACCEPTED {
		t.Logf("API should of returned ACCEPTED but didn't: %v", m.Message)
		t.FailNow()
	}
}

func TestApiStatsFail(t *testing.T) {
	srv := NewApiServer(c, s, tm, r, h, v, l)
	srv.setDefaultHandlers()

	req := httptest.NewRequest("GET", "http://127.0.0.1:9999/v1/api/stats?junkvalue", nil)
	w := httptest.NewRecorder()

	srv.stats(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m response.Deploy
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != response.FAILED {
		t.Logf("API should of returned FAILED but didn't: %v", m.Message)
		t.FailNow()
	}
}*/

func TestMain(m *testing.M) {
	srv := NewApiServer(c, s, tm, r, h, v, l)
	go srv.RunAPI(nil) // default handlers
	m.Run()
}
