package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/resources/manager"
	"mesos-framework-sdk/scheduler"
	"mesos-framework-sdk/server"
	taskbuilder "mesos-framework-sdk/task"
	sdkTaskManager "mesos-framework-sdk/task/manager"
	"net/http"
	"os"
	"sprint/scheduler/api/response"
	"sprint/task/builder"
	"strconv"
)

const (
	baseUrl        = "/v1/api"
	deployEndpoint = "/deploy"
	statusEndpoint = "/status"
	killEndpoint   = "/kill"
	updateEndpoint = "/update"
	statsEndpoint  = "/stats"
	retries        = 20
)

type ApiServer struct {
	cfg         server.Configuration
	mux         *http.ServeMux
	handle      map[string]http.HandlerFunc
	sched       scheduler.Scheduler
	taskMgr     sdkTaskManager.TaskManager
	resourceMgr manager.ResourceManager
	version     string
	logger      logging.Logger
}

func NewApiServer(
	cfg server.Configuration,
	s scheduler.Scheduler,
	t sdkTaskManager.TaskManager,
	r manager.ResourceManager,
	mux *http.ServeMux,
	version string,
	lgr logging.Logger) *ApiServer {

	return &ApiServer{
		cfg:         cfg,
		handle:      make(map[string]http.HandlerFunc),
		sched:       s,
		taskMgr:     t,
		resourceMgr: r,
		mux:         mux,
		version:     version,
		logger:      lgr,
	}
}

//Getter to return our map of handles
func (a *ApiServer) Handle() map[string]http.HandlerFunc {
	return a.handle
}

//Set our default API handler routes here.
func (a *ApiServer) setDefaultHandlers() {
	a.handle[baseUrl+deployEndpoint] = a.deploy
	a.handle[baseUrl+statusEndpoint] = a.state
	a.handle[baseUrl+killEndpoint] = a.kill
	a.handle[baseUrl+updateEndpoint] = a.update
	a.handle[baseUrl+statsEndpoint] = a.stats
}

func (a *ApiServer) setHandlers(handles map[string]http.HandlerFunc) {
	for route, handle := range handles {
		a.handle[route] = handle
	}
}

// RunAPI takes the scheduler controller and sets up the configuration for the API.
func (a *ApiServer) RunAPI(handlers map[string]http.HandlerFunc) {
	if handlers != nil || len(handlers) != 0 {
		a.logger.Emit(logging.INFO, "Setting custom handlers.")
		a.setHandlers(handlers)
	} else {
		a.logger.Emit(logging.INFO, "Setting default handlers.")
		a.setDefaultHandlers()
	}

	// Iterate through all methods and setup endpoints.
	for route, handle := range a.handle {
		a.mux.HandleFunc(route, handle)
	}

	if a.cfg.TLS() {
		a.cfg.Server().Handler = a.mux
		a.cfg.Server().Addr = ":" + strconv.Itoa(a.cfg.Port())
		if err := a.cfg.Server().ListenAndServeTLS(a.cfg.Cert(), a.cfg.Key()); err != nil {
			a.logger.Emit(logging.ERROR, err.Error())
			os.Exit(1)
		}
	} else {
		if err := http.ListenAndServe(":"+strconv.Itoa(a.cfg.Port()), a.mux); err != nil {
			a.logger.Emit(logging.ERROR, err.Error())
			os.Exit(1)
		}
	}
}

func (a *ApiServer) methodFilter(w http.ResponseWriter, r *http.Request, methods []string, ops func()) {
	for _, method := range methods {
		if method == r.Method {
			ops()
			return
		}
	}

	json.NewEncoder(w).Encode(response.Deploy{
		Status:  response.FAILED,
		Message: r.Method + " is not allowed on this endpoint.",
	})
}

// Deploys a given application from parsed JSON
func (a *ApiServer) deploy(w http.ResponseWriter, r *http.Request) {
	a.methodFilter(w, r, []string{"POST"}, func() {
		dec, err := ioutil.ReadAll(r.Body)
		if err != nil {
			json.NewEncoder(w).Encode(response.Deploy{
				Status:  response.FAILED,
				Message: err.Error(),
			})
			return
		}

		var m taskbuilder.ApplicationJSON
		err = json.Unmarshal(dec, &m)
		if err != nil {
			json.NewEncoder(w).Encode(response.Deploy{
				Status:  response.FAILED,
				Message: err.Error(),
			})
			return
		}

		task, err := builder.Application(&m, a.logger)
		if err != nil {
			json.NewEncoder(w).Encode(response.Deploy{
				Status:   response.FAILED,
				TaskName: task.GetName(),
				Message:  err.Error(),
			})
			return
		}

		// If we have any filters, let the resource manager know.
		if len(m.Filters) > 0 {
			if err := a.resourceMgr.AddFilter(task, m.Filters); err != nil {
				json.NewEncoder(w).Encode(response.Deploy{
					Status:   response.FAILED,
					TaskName: task.GetName(),
					Message:  err.Error(),
				})
				return
			}

		}

		if err := a.taskMgr.Add(task); err != nil {
			json.NewEncoder(w).Encode(response.Deploy{
				Status:   response.FAILED,
				TaskName: task.GetName(),
				Message:  err.Error(),
			})
			return
		}
		a.sched.Revive()

		json.NewEncoder(w).Encode(response.Deploy{
			Status:   response.QUEUED,
			TaskName: task.GetName(),
		})
	})
}

func (a *ApiServer) update(w http.ResponseWriter, r *http.Request) {
	a.methodFilter(w, r, []string{"PUT"}, func() {
		dec, err := ioutil.ReadAll(r.Body)
		if err != nil {
			json.NewEncoder(w).Encode(response.Deploy{
				Status:  response.FAILED,
				Message: err.Error(),
			})
			return
		}

		var m taskbuilder.ApplicationJSON
		err = json.Unmarshal(dec, &m)
		if err != nil {
			json.NewEncoder(w).Encode(response.Deploy{
				Status:  response.FAILED,
				Message: err.Error(),
			})
			return
		}

		// Check if this task already exists
		taskToKill, err := a.taskMgr.Get(&m.Name)
		if err != nil {
			json.NewEncoder(w).Encode(response.Deploy{
				Status:   response.FAILED,
				TaskName: taskToKill.GetName(),
				Message:  err.Error(),
			})
			return
		}

		task, err := builder.Application(&m, a.logger)

		a.taskMgr.Set(sdkTaskManager.UNKNOWN, task)
		a.sched.Kill(taskToKill.GetTaskId(), taskToKill.GetAgentId())
		a.sched.Revive()

		json.NewEncoder(w).Encode(response.Deploy{
			Status:  response.UPDATE,
			Message: fmt.Sprintf("Updating %v", task.GetName()),
		})
	})
}

func (a *ApiServer) kill(w http.ResponseWriter, r *http.Request) {
	a.methodFilter(w, r, []string{"DELETE"}, func() {
		dec, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}

		var m taskbuilder.KillJson
		err = json.Unmarshal(dec, &m)
		if err != nil {
			json.NewEncoder(w).Encode(response.Deploy{
				Status:  response.FAILED,
				Message: err.Error(),
			})
			return
		}

		// Make sure we have a name to look up
		if m.Name != nil {
			// Look up task in task manager
			t, err := a.taskMgr.Get(m.Name)
			if err != nil {
				json.NewEncoder(w).Encode(response.Kill{Status: response.NOTFOUND, TaskName: *m.Name})
				return
			}
			// Get all tasks in RUNNING state.
			running, _ := a.taskMgr.GetState(mesos_v1.TaskState_TASK_RUNNING)
			// If we get an error, it means no tasks are currently in the running state.
			// We safely ignore this- the range over the empty list will be skipped regardless.

			// Check if our task is in the list of RUNNING tasks.
			for _, task := range running {
				// If it is, then send the kill signal.
				if task.GetName() == t.GetName() {
					// First Kill call to the mesos-master.
					_, err := a.sched.Kill(t.GetTaskId(), t.GetAgentId())
					if err != nil {
						// If it fails, try to kill it again.
						resp, err := a.sched.Kill(t.GetTaskId(), t.GetAgentId())
						if err != nil {
							// We've tried twice and still failed.
							// Send back an error message.
							json.NewEncoder(w).Encode(
								response.Kill{
									Status:   response.FAILED,
									TaskName: *m.Name,
									Message:  "Response Status to Kill: " + resp.Status,
								},
							)
							return
						}
					}
					// Our kill call has worked, delete it from the task queue.
					a.taskMgr.Delete(t)
					a.resourceMgr.ClearFilters(t)
					// Response appropriately.
					json.NewEncoder(w).Encode(response.Kill{Status: response.KILLED, TaskName: *m.Name})
					return
				}
			}
			// If we get here, our task isn't in the list of RUNNING tasks.
			// Delete it from the queue regardless.
			// We run into this case if a task is flapping or unable to launch
			// or get an appropriate offer.
			a.taskMgr.Delete(t)
			a.resourceMgr.ClearFilters(t)
			json.NewEncoder(w).Encode(response.Kill{Status: response.KILLED, TaskName: *m.Name})
			return
		}
		// If we get here, there was no name passed in and the kill function failed.
		json.NewEncoder(w).Encode(response.Kill{Status: response.FAILED, TaskName: *m.Name})
	})
}

func (a *ApiServer) stats(w http.ResponseWriter, r *http.Request) {
	a.methodFilter(w, r, []string{"GET"}, func() {
		name := r.URL.Query().Get("name")

		t, err := a.taskMgr.Get(&name)
		if err != nil {
			json.NewEncoder(w).Encode(struct {
				Status   string
				TaskName string
				Message  string
			}{
				response.FAILED,
				name,
				"Failed to get state for task.",
			})
			return
		}
		task := a.taskMgr.Tasks().Get(t.GetName())
		json.NewEncoder(w).Encode(struct {
			Status   string
			TaskName string
			State    string
		}{
			response.ACCEPTED,
			t.GetName(),
			task.(sdkTaskManager.Task).State.String(),
		})
		return
	})
}

// Status endpoint lets the end-user know about the TASK_STATUS of their task.
func (a *ApiServer) state(w http.ResponseWriter, r *http.Request) {
	a.methodFilter(w, r, []string{"GET"}, func() {
		name := r.URL.Query().Get("name")

		_, err := a.taskMgr.Get(&name)
		if err != nil {
			json.NewEncoder(w).Encode(response.Deploy{
				Status:  response.FAILED,
				Message: err.Error(),
			})
			return
		}
		queued, err := a.taskMgr.GetState(sdkTaskManager.STAGING)
		if err != nil {
			a.logger.Emit(logging.INFO, err.Error())
		}

		for _, task := range queued {
			if task.GetName() == name {
				json.NewEncoder(w).Encode(response.Kill{Status: response.QUEUED, TaskName: name})
			}
		}

		json.NewEncoder(w).Encode(response.Kill{Status: response.LAUNCHED, TaskName: name})
	})
}
