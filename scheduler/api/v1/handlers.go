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
	apiManager "hydrogen/scheduler/api/manager"
	"io/ioutil"
	"mesos-framework-sdk/include/mesos_v1"
	"net/http"
)

// API handlers communicate with the API manager to perform the appropriate actions.
type Handlers struct {
	manager apiManager.ApiParser
}

// Returns a new handlers instance for mapping routes.
func NewHandlers(mgr apiManager.ApiParser) *Handlers {
	return &Handlers{manager: mgr}
}

// Deploy handler launches a given application from parsed JSON.
func (h *Handlers) Application(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.deployApplication(w, r)
	case http.MethodDelete:
		h.killApplication(w, r)
	case http.MethodPut:
		h.updateApplication(w, r)
	case http.MethodGet:
		h.applicationState(w, r)
	default:
		MethodNotAllowed(w, MessageResponse{r.Method + " is not allowed on this endpoint"})
	}
}

// Deploys an application.
func (h *Handlers) deployApplication(w http.ResponseWriter, r *http.Request) {
	dec, err := ioutil.ReadAll(r.Body)
	if err != nil {
		BadRequest(w, MessageResponse{err.Error()})
		return
	}

	defer r.Body.Close()

	tasks, err := h.manager.Deploy(dec)
	if err != nil {
		InternalServerError(w, MessageResponse{err.Error()})
		return
	}

	resp := []MessageResponse{}
	for _, task := range tasks {
		resp = append(resp, MessageResponse{
			"Task " + task.Info.GetName() + " successfully queued",
		})
	}

	Success(w, resp)
}

// Update handler allows for updates to an existing/running task.
func (h *Handlers) updateApplication(w http.ResponseWriter, r *http.Request) {
	dec, err := ioutil.ReadAll(r.Body)
	if err != nil {
		BadRequest(w, MessageResponse{err.Error()})
		return
	}

	defer r.Body.Close()

	newTasks, err := h.manager.Update(dec)
	if err != nil {
		InternalServerError(w, MessageResponse{err.Error()})
		return
	}

	resp := []MessageResponse{}
	for _, task := range newTasks {
		resp = append(resp, MessageResponse{
			"Task " + task.Info.GetName() + " successfully queued",
		})
	}

	Success(w, resp)
}

// Kill handler allows users to stop their running task.
func (h *Handlers) killApplication(w http.ResponseWriter, r *http.Request) {
	dec, err := ioutil.ReadAll(r.Body)
	if err != nil {
		BadRequest(w, MessageResponse{err.Error()})
		return
	}

	defer r.Body.Close()

	name, err := h.manager.Kill(dec)
	if err != nil {
		BadRequest(w, MessageResponse{err.Error()})
		return
	}

	Success(w, MessageResponse{
		Message: "Successfully killed task task " + name,
	})
}

// State handler provides the given task's current execution status.
func (h *Handlers) applicationState(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		BadRequest(w, MessageResponse{"Task name is required"})
		return
	}

	task, err := h.manager.Status(name)
	if err != nil {
		InternalServerError(w, MessageResponse{err.Error()})
		return
	}

	Success(w, struct {
		Name      *string                 `json:"name"`
		State     string                  `json:"state"`
		Instances int                     `json:"instances"`
		Resources []*mesos_v1.Resource    `json:"resources,omitempty"`
		Container *mesos_v1.ContainerInfo `json:"container,omitempty"`
		Command   *mesos_v1.CommandInfo   `json:"command,omitempty"`
		Labels    *mesos_v1.Labels        `json:"labels,omitempty"`
		Executor  *mesos_v1.ExecutorInfo  `json:"executor,omitempty"`
	}{
		task.Info.Name,
		task.State.String(),
		task.Instances,
		task.Info.Resources,
		task.Info.Container,
		task.Info.Command,
		task.Info.Labels,
		task.Info.Executor,
	})
}

// Tasks handler provides a list of all tasks known to the scheduler.
func (h *Handlers) Tasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getAllTasks(w, r)
	default:
		MethodNotAllowed(w, MessageResponse{r.Method + " is not allowed on this endpoint"})
	}
}

// Gathers all tasks known to the scheduler.
func (h *Handlers) getAllTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.manager.AllTasks()
	if err != nil {
		// This isn't an error since it's expected the task manager can be empty.
		Success(w, MessageResponse{err.Error()})
		return
	}

	resp := []struct {
		Name  string
		State string
	}{}

	for _, t := range tasks {
		resp = append(resp, struct {
			Name  string
			State string
		}{
			t.Info.GetName(),
			t.State.String(),
		})
	}

	Success(w, resp)
}
