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

// Helper for sending responses indicating failure.
func failureResponse(w http.ResponseWriter, msg string, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Response{
		Status:  FAILED,
		Message: msg,
	})
}

// Helper for sending responses indicating success.
func successResponse(w http.ResponseWriter, status, name, msg string) {
	json.NewEncoder(w).Encode(Response{
		Status:   status,
		TaskName: name,
		Message:  msg,
	})
}

// Deploy handler launches a given application from parsed JSON.
func (h *Handlers) Deploy(w http.ResponseWriter, r *http.Request) {
	dec, err := ioutil.ReadAll(r.Body)
	if err != nil {
		failureResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	task, err := h.manager.Deploy(dec)
	if err != nil {
		failureResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	successResponse(w, QUEUED, task.GetName(), "Task successfully queued.")
}

// Update handler allows for updates to an existing/running task.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	dec, err := ioutil.ReadAll(r.Body)
	if err != nil {
		failureResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	newTask, err := h.manager.Update(dec)
	if err != nil {
		failureResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	name := newTask.GetName()

	successResponse(w, UPDATE, name, "Updating "+name+".")
}

// Kill handler allows users to stop their running task.
func (h *Handlers) Kill(w http.ResponseWriter, r *http.Request) {
	dec, err := ioutil.ReadAll(r.Body)
	if err != nil {
		failureResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	name, err := h.manager.Kill(dec)
	if err != nil {
		failureResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	successResponse(w, KILLED, name, "Successfully killed task task "+name)
}

// State handler provides the given task's current execution status.
func (h *Handlers) State(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		failureResponse(w, "No name was found in URL params.", http.StatusBadRequest)
		return
	}

	state, err := h.manager.Status(name)
	if err != nil {
		failureResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stateStr := state.String()

	successResponse(w, stateStr, name, "Task is "+stateStr)
}

// Tasks handler provides a list of all tasks known to the scheduler.
func (h *Handlers) Tasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.manager.AllTasks()
	if err != nil {
		failureResponse(w, err.Error(), http.StatusInternalServerError)
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
