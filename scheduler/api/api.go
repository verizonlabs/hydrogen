package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/server"
	taskbuilder "mesos-framework-sdk/task"
	"net/http"
	"os"
	"sprint/scheduler/api/response"
	"sprint/scheduler/events"
	"sprint/task/builder"
	"strconv"
	"time"
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
	cfg       server.Configuration
	port      *int
	mux       *http.ServeMux
	handle    map[string]http.HandlerFunc // route -> handler func for that route
	eventCtrl *events.SprintEventController
	version   string
	logger    logging.Logger
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
	a.handle = make(map[string]http.HandlerFunc, 4)
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

func (a *ApiServer) setEventController(e *events.SprintEventController) {
	a.eventCtrl = e
}

// RunAPI takes the scheduler controller and sets up the configuration for the API.
func (a *ApiServer) RunAPI(e *events.SprintEventController, handlers map[string]http.HandlerFunc) {
	if handlers != nil || len(handlers) != 0 {
		a.logger.Emit(logging.INFO, "Setting custom handlers.")
		a.setHandlers(handlers)
	} else {
		a.logger.Emit(logging.INFO, "Setting default handlers.")
		a.setDefaultHandlers()
	}

	a.setEventController(e)

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
				a.eventCtrl.ResourceManager().AddFilter(task, m.Filters)
			}

			a.eventCtrl.TaskManager().Add(task)
			a.eventCtrl.Scheduler().Revive()

			fmt.Fprintf(w, "%v", task)
		}
	default:
		{
			fmt.Fprintf(w, r.Method+" is not allowed on this endpoint.")
		}
	}

}

func (a *ApiServer) update(w http.ResponseWriter, r *http.Request) {
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

			// Check if this task already exists
			taskToKill, err := a.eventCtrl.TaskManager().Get(&m.Name)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				return
			}

			task, err := builder.Application(&m, a.logger)

			a.eventCtrl.TaskManager().Add(task)
			a.eventCtrl.Scheduler().Revive()

			go func() {
				for i := 0; i < retries; i++ {
					launched := a.eventCtrl.TaskManager().LaunchedTasks()
					if _, ok := launched[task.GetName()]; ok {
						a.eventCtrl.TaskManager().Delete(taskToKill)
						a.eventCtrl.Scheduler().Kill(taskToKill.GetTaskId(), taskToKill.GetAgentId())
						return
					}
					time.Sleep(2 * time.Second) // Wait a pre-determined amount of time for polling.
				}
				return
			}()
			fmt.Fprintf(w, "Updating %v", task.GetName())
		}
	default:
		{
			fmt.Fprintf(w, r.Method+" is not allowed on this endpoint.")
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

			var status string
			if m.Name != nil {
				t, err := a.eventCtrl.TaskManager().Get(m.Name)
				if err != nil {
					json.NewEncoder(w).Encode(response.Kill{Status: response.NOTFOUND, TaskName: *m.Name})
				} else {
					fmt.Println("In kill.")
					a.eventCtrl.TaskManager().Delete(t)
					a.eventCtrl.Scheduler().Kill(t.GetTaskId(), t.GetAgentId())
					json.NewEncoder(w).Encode(response.Kill{Status: response.KILLED, TaskName: *m.Name})
				}
			} else {
				json.NewEncoder(w).Encode(response.Kill{Status: response.FAILED, TaskName: *m.Name})
			}
			fmt.Fprintf(w, "%v", status)
		}
	default:
		{
			fmt.Fprintf(w, r.Method+" is not allowed on this endpoint.")
		}
	}

}

func (a *ApiServer) stats(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		{
			name := r.URL.Query().Get("name")

			_, err := a.eventCtrl.TaskManager().Get(&name)
			if err != nil {
				fmt.Fprintf(w, "Task not found, error %v", err.Error())
				return
			}
			// get the task from task queue to
		}
	default:
		{
			fmt.Fprintf(w, r.Method+" is not allowed on this endpoint.")
		}
	}
}

// Status endpoint lets the end-user know about the TASK_STATUS of their task.
func (a *ApiServer) state(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		{
			name := r.URL.Query().Get("name")

			t, err := a.eventCtrl.TaskManager().Get(&name)
			if err != nil {
				fmt.Fprintf(w, "Task not found, error %v", err.Error())
				return
			}
			queued := a.eventCtrl.TaskManager().QueuedTasks()
			var status string
			if task, ok := queued[t.GetTaskId().GetValue()]; ok {
				json.NewEncoder(w).Encode(response.Kill{Status: response.LAUNCHED, TaskName: task.GetName()})
			} else {
				json.NewEncoder(w).Encode(response.Kill{Status: response.QUEUED, TaskName: task.GetName()})
			}
			fmt.Fprintf(w, "%v", status)

		}
	default:
		{
			fmt.Fprintf(w, r.Method+" is not allowed on this endpoint.")
		}
	}

}
