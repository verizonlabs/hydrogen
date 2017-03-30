package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	port        *int
	mux         *http.ServeMux
	handle      map[string]http.HandlerFunc
	sched       scheduler.Scheduler
	taskMgr     sdkTaskManager.TaskManager
	resourceMgr manager.ResourceManager
	version     string
	logger      logging.Logger
}

func NewApiServer(cfg server.Configuration, mux *http.ServeMux, port *int, version string, lgr *logging.DefaultLogger) *ApiServer {
	return &ApiServer{
		cfg:     cfg,
		port:    port,
		mux:     mux,
		version: version,
		logger:  lgr,
	}
}

//Getter to return our map of handles
func (a *ApiServer) Handle() map[string]http.HandlerFunc {
	return a.handle
}

//Set our default API handler routes here.
func (a *ApiServer) setDefaultHandlers() {
	a.handle = make(map[string]http.HandlerFunc, 5)
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
func (a *ApiServer) RunAPI(s scheduler.Scheduler, t sdkTaskManager.TaskManager, r manager.ResourceManager, handlers map[string]http.HandlerFunc) {
	if handlers != nil || len(handlers) != 0 {
		a.logger.Emit(logging.INFO, "Setting custom handlers.")
		a.setHandlers(handlers)
	} else {
		a.logger.Emit(logging.INFO, "Setting default handlers.")
		a.setDefaultHandlers()
	}

	a.sched = s
	a.taskMgr = t
	a.resourceMgr = r

	// Iterate through all methods and setup endpoints.
	for route, handle := range a.handle {
		a.mux.HandleFunc(route, handle)
	}

	if a.cfg.TLS() {
		a.cfg.Server().Handler = a.mux
		a.cfg.Server().Addr = ":" + strconv.Itoa(*a.port)
		if err := a.cfg.Server().ListenAndServeTLS(a.cfg.Cert(), a.cfg.Key()); err != nil {
			a.logger.Emit(logging.ERROR, err.Error())
			os.Exit(1)
		}
	} else {
		if err := http.ListenAndServe(":"+strconv.Itoa(*a.port), a.mux); err != nil {
			a.logger.Emit(logging.ERROR, err.Error())
			os.Exit(1)
		}
	}
}

// Deploys a given application from parsed JSON
func (a *ApiServer) deploy(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		{
			dec, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				return
			}

			var m taskbuilder.ApplicationJSON
			err = json.Unmarshal(dec, &m)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				return
			}

			task, err := builder.Application(&m, a.logger)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				return
			}

			// If we have any filters, let the resource manager know.
			if len(m.Filters) > 0 {
				a.resourceMgr.AddFilter(task, m.Filters)
			}

			if err := a.taskMgr.Add(task); err != nil {
				fmt.Fprintln(w, err.Error())
				return
			}
			a.sched.Revive()

			json.NewEncoder(w).Encode(response.Deploy{
				Status:   response.QUEUED,
				TaskName: task.GetName(),
			})
		}
	default:
		{
			json.NewEncoder(w).Encode(response.Deploy{
				Status:  response.FAILED,
				Message: r.Method + " is not allowed on this endpoint.",
			})
		}
	}

}

func (a *ApiServer) update(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		{
			dec, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				return
			}

			var m taskbuilder.ApplicationJSON
			err = json.Unmarshal(dec, &m)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				return
			}

			// Check if this task already exists
			taskToKill, err := a.taskMgr.Get(&m.Name)
			if err != nil {
				fmt.Fprintf(w, err.Error())
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
		}
	default:
		{
			json.NewEncoder(w).Encode(response.Deploy{
				Status:  response.FAILED,
				Message: r.Method + " is not allowed on this endpoint.",
			})
		}
	}
}

func (a *ApiServer) kill(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		{
			dec, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return
			}

			var m taskbuilder.KillJson
			err = json.Unmarshal(dec, &m)
			if err != nil {
				fmt.Fprintf(w, "%v", err.Error())
				return
			}

			if m.Name != nil {
				t, err := a.taskMgr.Get(m.Name)
				if err != nil {
					json.NewEncoder(w).Encode(response.Kill{Status: response.NOTFOUND, TaskName: *m.Name})
				} else {
					a.sched.Kill(t.GetTaskId(), t.GetAgentId()) // Check to see that it's killed...
					a.taskMgr.Delete(t)
					json.NewEncoder(w).Encode(response.Kill{Status: response.KILLED, TaskName: *m.Name})
				}
			} else {
				json.NewEncoder(w).Encode(response.Kill{Status: response.FAILED, TaskName: *m.Name})
			}
		}
	default:
		{
			json.NewEncoder(w).Encode(response.Deploy{
				Status:  response.FAILED,
				Message: r.Method + " is not allowed on this endpoint.",
			})
		}
	}

}

func (a *ApiServer) stats(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		{
			name := r.URL.Query().Get("name")

			_, err := a.taskMgr.Get(&name)
			if err != nil {
				fmt.Fprintf(w, "Task not found, error %v", err.Error())
				return
			}
			// get the task from task queue to
		}
	default:
		{
			json.NewEncoder(w).Encode(response.Deploy{
				Status:  response.FAILED,
				Message: r.Method + " is not allowed on this endpoint.",
			})
		}
	}
}

// Status endpoint lets the end-user know about the TASK_STATUS of their task.
func (a *ApiServer) state(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		{
			name := r.URL.Query().Get("name")

			_, err := a.taskMgr.Get(&name)
			if err != nil {
				fmt.Fprintf(w, "Task not found, error %v", err.Error())
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

		}
	default:
		{
			fmt.Fprintf(w, r.Method+" is not allowed on this endpoint.")
		}
	}

}
