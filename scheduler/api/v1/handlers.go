package v1

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	apiManager "sprint/scheduler/api/manager"
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
func (h *Handlers) Deploy(w http.ResponseWriter, r *http.Request) {
	dec, err := ioutil.ReadAll(r.Body)
	if err != nil {
		BadRequest(w, Response{Message: err.Error()})
		return
	}

	defer r.Body.Close()

	task, err := h.manager.Deploy(dec)
	if err != nil {
		InternalServerError(w, Response{Message: err.Error()})
		return
	}

	Success(w, Response{TaskName: task.GetName(), Message: "Task successfully queued."})
}

// Update handler allows for updates to an existing/running task.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	dec, err := ioutil.ReadAll(r.Body)
	if err != nil {
		BadRequest(w, Response{Message: err.Error()})
		return
	}

	defer r.Body.Close()

	newTask, err := h.manager.Update(dec)
	if err != nil {
		InternalServerError(w, Response{Message: err.Error()})
		return
	}

	name := newTask.GetName()

	Success(w, Response{Message: "Updating " + name + ".", TaskName: newTask.GetName()})
}

// Kill handler allows users to stop their running task.
func (h *Handlers) Kill(w http.ResponseWriter, r *http.Request) {
	dec, err := ioutil.ReadAll(r.Body)
	if err != nil {
		BadRequest(w, Response{Message: err.Error()})
		return
	}

	defer r.Body.Close()

	name, err := h.manager.Kill(dec)
	if err != nil {
		InternalServerError(w, Response{Message: err.Error()})
		return
	}

	Success(w, Response{Message: "Successfully killed task task " + name})
}

// State handler provides the given task's current execution status.
func (h *Handlers) State(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		BadRequest(w, Response{Message: "No name was found in URL params."})
		return
	}

	state, err := h.manager.Status(name)
	if err != nil {
		InternalServerError(w, Response{Message: err.Error()})
		return
	}

	Success(w, Response{Message: "Task is " + state.String()})
}

// Tasks handler provides a list of all tasks known to the scheduler.
func (h *Handlers) Tasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.manager.AllTasks()
	if err != nil {
		// This isn't an error since it's expected the task manager can be empty.
		Success(w, Response{Message: err.Error()})
		return
	}

	data := []Response{}
	for _, t := range tasks {
		data = append(data, Response{
			State:    t.State.String(),
			TaskName: t.Info.GetName(),
		})
	}

	json.NewEncoder(w).Encode(data)
}
