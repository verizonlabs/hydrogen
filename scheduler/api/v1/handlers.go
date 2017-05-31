package v1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	apiManager "sprint/scheduler/api/manager"
)

// API server used for scheduling/updating/killing tasks.
// Provides an interface for users to interact with the core scheduler.
type Handlers struct {
	manager apiManager.ApiParser
}

// Returns a new API server injected with the necessary components.
func NewHandlers(mgr apiManager.ApiParser) *Handlers {
	return &Handlers{manager: mgr}
}

// HTTP method guard to only allow specific methods on various endpoints.
func (a *Handlers) methodFilter(w http.ResponseWriter, r *http.Request, methods []string, ops func()) {
	for _, method := range methods {
		if method == r.Method {
			ops()
			return
		}
	}

	json.NewEncoder(w).Encode(Response{
		Status:  FAILED,
		Message: r.Method + " is not allowed on this endpoint.",
	})
}

// Deploy handler launches a given application from parsed JSON.
func (h *Handlers) Deploy(w http.ResponseWriter, r *http.Request) {
	h.methodFilter(w, r, []string{"POST"}, func() {
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
			json.NewEncoder(w).Encode(Response{
				Status:  FAILED,
				Message: err.Error(),
			})
			return
		}

		json.NewEncoder(w).Encode(Response{
			Status:   QUEUED,
			TaskName: task.GetName(),
		})
	})
}

// Update handler allows for updates to an existing/running task.
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	h.methodFilter(w, r, []string{"PUT"}, func() {
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
			json.NewEncoder(w).Encode(Response{
				Status:  FAILED,
				Message: err.Error(),
			})
			return
		}

		json.NewEncoder(w).Encode(Response{
			Status:  UPDATE,
			Message: fmt.Sprintf("Updating %v", newTask.GetName()),
		})
	})
}

// Kill handler allows users to stop their running task.
func (h *Handlers) Kill(w http.ResponseWriter, r *http.Request) {
	h.methodFilter(w, r, []string{"DELETE"}, func() {
		dec, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}

		defer r.Body.Close()

		err = h.manager.Kill(dec)
		if err != nil {
			json.NewEncoder(w).Encode(Response{
				Status:  FAILED,
				Message: err.Error(),
			})
			return
		}
		json.NewEncoder(w).Encode(Response{
			Status:  KILLED,
			Message: "Successfully killed task.",
		})
	})
}

// State handler provides the given task's current execution status.
func (h *Handlers) State(w http.ResponseWriter, r *http.Request) {
	h.methodFilter(w, r, []string{"GET"}, func() {
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
			json.NewEncoder(w).Encode(Response{
				Status:  FAILED,
				Message: err.Error(),
			})
			return
		}

		json.NewEncoder(w).Encode(Response{Status: state.String(), TaskName: name})
	})
}

// Tasks handler provides a list of all tasks known to the scheduler.
func (h *Handlers) Tasks(w http.ResponseWriter, r *http.Request) {
	h.methodFilter(w, r, []string{"GET"}, func() {
		tasks, err := h.manager.AllTasks()
		if err != nil {
			json.NewEncoder(w).Encode(Response{
				Status:  FAILED,
				Message: err.Error(),
			})
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
	})
}
