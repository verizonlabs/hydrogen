package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mesos-framework-sdk/logging/test"
	"mesos-framework-sdk/resources/manager/test"
	"mesos-framework-sdk/scheduler/test"
	"mesos-framework-sdk/server/test"
	"mesos-framework-sdk/task/manager/test"
	"net/http"
	"net/http/httptest"
	"sprint/scheduler/api/manager"
	"strings"
	"testing"
)

type brokenReader struct{}

func (b brokenReader) Read(n []byte) (int, error) {
	return 0, errors.New("I'm broke.")
}

var (
	c                     = new(mockConfiguration.MockServerConfiguration)
	s                     = new(scheduler.MockScheduler)
	tm                    = new(testTaskManager.MockTaskManager)
	r                     = new(MockResourceManager.MockResourceManager)
	h                     = http.NewServeMux()
	v                     = "test"
	l                     = new(MockLogging.MockLogger)
	apiMgr                = new(apimanager.MockApiManager)
	apiBrokenMgr          = new(apimanager.MockBrokenApiManager)
	validJSON             = fmt.Sprint(`{"name": "test", "resources": {"cpu": 0.5, "mem": 128.0}, "command": {"cmd": "echo hello"}}`)
	killJSON              = fmt.Sprint(`{"name": "test"}`)
	junkJSON              = fmt.Sprint(`not even json, how did this even get here`)
	filtersJSON           = fmt.Sprint(`{"name": "test", "filters": [{"type": "TEXT", "value": ["tester"]}], "resources": {"cpu": 0.5, "mem": 128.0}, "command": {"cmd": "echo hello"}}`)
	badFiltersJSON        = fmt.Sprint(`{"name": "test", "filters": [{"type": "not real", "value": "tester"}], "resources": {"cpu": 0.5, "mem": 128.0}, "command": {"cmd": "echo hello"}}`)
	invalidFilterTypeJSON = fmt.Sprint(`{"name": "test", "filters": [{"type": "fake news", "value": ["tester"]}], "resources": {"cpu": 0.5, "mem": 128.0}, "command": {"cmd": "echo hello"}}`)
)

// Ensures all components are set correctly when creating the API server.
func TestNewApiServer(t *testing.T) {
	srv := NewApiServer(c, apiMgr, h, v, l)
	if srv.cfg != c || srv.mux != h || srv.version != v || srv.logger != l {
		t.Fatal("API does not contain the correct components")
	}
}

func TestNewApiServerRun(t *testing.T) {
	srv := NewApiServer(c, apiMgr, h, v, l)
	if srv.cfg != c || srv.mux != h || srv.version != v || srv.logger != l {
		t.Fatal("API does not contain the correct components")
	}
}

func TestApiServer_Handle(t *testing.T) {
	srv := NewApiServer(c, apiMgr, h, v, l)
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
	srv := NewApiServer(c, apiMgr, h, v, l)
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

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status == FAILED {
		t.Logf("Task shouldn't of failed but did %v", m.Message)
		t.FailNow()
	}
}

func TestApiJunkDeploy(t *testing.T) {
	srv := NewApiServer(c, apiBrokenMgr, h, v, l)
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

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != FAILED {
		t.FailNow()
	}
}

func TestApiDeployBrokenTask(t *testing.T) {
	srv := NewApiServer(c, apiBrokenMgr, h, v, l)
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

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != FAILED {
		t.Logf("Task shouldn of failed but didn't %v", m.Message)
		t.FailNow()
	}
}

func TestApiUpdate(t *testing.T) {
	srv := NewApiServer(c, apiMgr, h, v, l)
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

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != UPDATE {
		t.Logf("Task should of been UPDATE but didn't %v", m.Message)
		t.FailNow()
	}
}

func TestApiUpdateFail(t *testing.T) {
	srv := NewApiServer(c, apiBrokenMgr, h, v, l)
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

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != FAILED {
		t.Logf("Task should of been FAILED but wasn't %v", m.Message)
		t.FailNow()
	}
}

func TestApiUpdateFailTask(t *testing.T) {
	srv := NewApiServer(c, apiBrokenMgr, h, v, l)
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

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != FAILED {
		t.Logf("Task should of been FAILED but wasn't %v", m.Message)
		t.FailNow()
	}
}

func TestApiKill(t *testing.T) {
	srv := NewApiServer(c, apiMgr, h, v, l)
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

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != KILLED {
		t.Logf("Task shouldn't of failed but did %v", m.Message)
		t.FailNow()
	}
}

func TestApiKillFail(t *testing.T) {
	srv := NewApiServer(c, apiBrokenMgr, h, v, l)
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

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != FAILED {
		t.Logf("Task should of failed but didn't:  %v", m.Message)
		t.FailNow()
	}
}

func TestApiKillFailTask(t *testing.T) {
	srv := NewApiServer(c, apiBrokenMgr, h, v, l)
	srv.setDefaultHandlers()

	a := strings.NewReader(validJSON)
	req := httptest.NewRequest("DELETE", "http://127.0.0.1:9999/v1/api/kill", a)
	w := httptest.NewRecorder()

	srv.kill(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != FAILED {
		t.Logf("Task should of been NOTFOUND but wasn't %v", m.Message)
		t.FailNow()
	}
}

func TestApiKillFailSchedulerCall(t *testing.T) {
	srv := NewApiServer(c, apiBrokenMgr, h, v, l)
	srv.setDefaultHandlers()

	a := strings.NewReader(validJSON)
	req := httptest.NewRequest("DELETE", "http://127.0.0.1:9999/v1/api/kill", a)
	w := httptest.NewRecorder()

	srv.kill(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != FAILED {
		t.Logf("Task should of been FAILED but wasn't %v", m.Message)
		t.FailNow()
	}
}

func TestApiState(t *testing.T) {
	srv := NewApiServer(c, apiMgr, h, v, l)
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

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != RUNNING {
		t.Logf("Task should of been in state RUNNING but wasn't:  %v", m.Message)
		t.FailNow()
	}
}

func TestApiStateFail(t *testing.T) {
	srv := NewApiServer(c, apiMgr, h, v, l)
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

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != FAILED {
		t.Logf("API should of returned FAILED but didn't:  %v", m.Message)
		t.FailNow()
	}
}

func TestApiStateFailTask(t *testing.T) {
	srv := NewApiServer(c, apiBrokenMgr, h, v, l)
	srv.setDefaultHandlers()

	req := httptest.NewRequest("GET", "http://127.0.0.1:9999/v1/api/status?name=junk", nil)
	w := httptest.NewRecorder()

	srv.state(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != FAILED {
		t.Logf("API should of returned FAILED but didn't:  %v", m.Message)
		t.FailNow()
	}
}

func TestApiFailWrongMethod(t *testing.T) {
	srv := NewApiServer(c, apiMgr, h, v, l)
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

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != FAILED {
		t.Logf("API should of returned FAILED but didn't:  %v", m.Message)
		t.FailNow()
	}
}

func TestApiDeployWithNil(t *testing.T) {
	srv := NewApiServer(c, apiBrokenMgr, h, v, l)
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

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != FAILED {
		t.Log(m)
		t.Logf("Task shouldn't of failed but did %v", m.Message)
		t.FailNow()
	}
}

func TestApiDeployWithFilters(t *testing.T) {
	srv := NewApiServer(c, apiMgr, h, v, l)
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

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status == FAILED {
		t.Logf("Task shouldn't of failed but did %v", m.Message)
		t.FailNow()
	}
}

func TestApiDeployWithFiltersFail(t *testing.T) {
	srv := NewApiServer(c, apiBrokenMgr, h, v, l)
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

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != FAILED {
		t.Logf("Task should of failed but didn't %v", m.Message)
		t.FailNow()
	}
}

func TestApiDeployWithIoutilFail(t *testing.T) {
	srv := NewApiServer(c, apiMgr, h, v, l)
	srv.setDefaultHandlers()

	a := brokenReader{}
	req := httptest.NewRequest("POST", "http://127.0.0.1:9999/v1/api/deploy", a)
	w := httptest.NewRecorder()

	srv.deploy(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != FAILED {
		t.Logf("Task should of failed but didn't %v", m.Message)
		t.FailNow()
	}
}

func TestApiDeployWithIoutilFilterFail(t *testing.T) {
	srv := NewApiServer(c, apiBrokenMgr, h, v, l)
	srv.setDefaultHandlers()

	a := strings.NewReader(invalidFilterTypeJSON)
	req := httptest.NewRequest("POST", "http://127.0.0.1:9999/v1/api/deploy", a)
	w := httptest.NewRecorder()

	srv.deploy(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != FAILED {
		t.Logf("Task should of failed but didn't %v", m.Message)
		t.FailNow()
	}
}

func TestApiUpdateWithIoutilFail(t *testing.T) {
	srv := NewApiServer(c, apiMgr, h, v, l)
	srv.setDefaultHandlers()

	a := brokenReader{}
	req := httptest.NewRequest("PUT", "http://127.0.0.1:9999/v1/api/update", a)
	w := httptest.NewRecorder()

	srv.update(w, req)

	resp := w.Result()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Logf("Error %v", err.Error())
		t.FailNow()
	}

	var m Response
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Logf("error unmarshalling %v", err)
		t.FailNow()
	}

	if m.Status != FAILED {
		t.Logf("Task should of failed but didn't %v", m.Message)
		t.FailNow()
	}
}

func TestMain(m *testing.M) {
	srv := NewApiServer(c, apiMgr, h, v, l)
	go srv.RunAPI(nil) // default handlers
	m.Run()
}
