package v1

import (
	"encoding/json"
	"fmt"
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

func failureResponse(w http.ResponseWriter, msg string) {
	json.NewEncoder(w).Encode(Response{
		Status:  FAILED,
		Message: msg,
	})
}

// Deploy handler launches a given application from parsed JSON.
func (h *Handlers) Deploy(w http.ResponseWriter, r *http.Request) {
	dec, err := ioutil.ReadAll(r.Body)
	if err != nil {
		json.NewEncoder(w).Encode(Response{
			Status:  FAILED,
			Message: err.Error(),
		})
		return
	}

	defer r.Body.Close()

	task, err := h.manager.Deploy(dec)
	if err != nil {
		failureResponse(w, err.Error())
		return
	}

	json.NewEncoder(w).Encode(Response{
		Status:   QUEUED,
		TaskName: task.GetName(),
	})
}

// Update handler allows for updates to an existing/running task.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	dec, err := ioutil.ReadAll(r.Body)
	if err != nil {
		json.NewEncoder(w).Encode(Response{
			Status:  FAILED,
			Message: err.Error(),
		})
		return
	}

	defer r.Body.Close()

	newTask, err := h.manager.Update(dec)
	if err != nil {
		failureResponse(w, err.Error())
		return
	}

	json.NewEncoder(w).Encode(Response{
		Status:  UPDATE,
		Message: fmt.Sprintf("Updating %v", newTask.GetName()),
	})
}

// Kill handler allows users to stop their running task.
func (h *Handlers) Kill(w http.ResponseWriter, r *http.Request) {
	dec, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	defer r.Body.Close()

	err = h.manager.Kill(dec)
	if err != nil {
		failureResponse(w, err.Error())
		return
	}

	json.NewEncoder(w).Encode(Response{
		Status:  KILLED,
		Message: "Successfully killed task.",
	})
}

// State handler provides the given task's current execution status.
func (h *Handlers) State(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		json.NewEncoder(w).Encode(Response{
			Status:  FAILED,
			Message: "No name was found in URL params.",
		})
		return
	}
	state, err := h.manager.Status(name)
	if err != nil {
		failureResponse(w, err.Error())
		return
	}

	json.NewEncoder(w).Encode(Response{Status: state.String(), TaskName: name})
}

// Tasks handler provides a list of all tasks known to the scheduler.
func (h *Handlers) Tasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.manager.AllTasks()
	if err != nil {
		failureResponse(w, err.Error())
		return
	}

	data := []Response{}
	for _, t := range tasks {
		data = append(data, Response{
			Status:   t.State.String(),
			TaskName: t.Info.GetName(),
		})
	}

	json.NewEncoder(w).Encode(data)
}
