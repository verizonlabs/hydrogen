package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/scheduler"
	"mesos-framework-sdk/server"
	"net/http"
	"os"
	apiManager "sprint/scheduler/api/manager"
	"strconv"
)

const (
	baseUrl        = "/v1/api"
	deployEndpoint = "/deploy"
	statusEndpoint = "/status"
	tasksEndpoint  = "/tasks"
	killEndpoint   = "/kill"
	updateEndpoint = "/update"

	ACCEPTED = "Accepted"
	LAUNCHED = "Launched"
	FAILED   = "Failed"
	RUNNING  = "Running"
	KILLED   = "Killed"
	NOTFOUND = "Not Found"
	QUEUED   = "Queued"
	UPDATE   = "Updated"
)

type Response struct {
	Status   string
	TaskName string
	Message  string
}

type ApiServer struct {
	cfg     server.Configuration
	mux     *http.ServeMux
	handle  map[string]http.HandlerFunc
	sched   scheduler.Scheduler
	manager apiManager.ApiManager
	version string
	logger  logging.Logger
}

func NewApiServer(
	cfg server.Configuration,
	mgr apiManager.ApiManager,
	mux *http.ServeMux,
	version string,
	lgr logging.Logger) *ApiServer {

	return &ApiServer{
		cfg:     cfg,
		handle:  make(map[string]http.HandlerFunc),
		manager: mgr,
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
	a.handle[baseUrl+deployEndpoint] = a.deploy
	a.handle[baseUrl+statusEndpoint] = a.state
	a.handle[baseUrl+killEndpoint] = a.kill
	a.handle[baseUrl+updateEndpoint] = a.update
	a.handle[baseUrl+tasksEndpoint] = a.tasks
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

	json.NewEncoder(w).Encode(Response{
		Status:  FAILED,
		Message: r.Method + " is not allowed on this endpoint.",
	})
}

// Deploys a given application from parsed JSON
func (a *ApiServer) deploy(w http.ResponseWriter, r *http.Request) {
	a.methodFilter(w, r, []string{"POST"}, func() {
		dec, err := ioutil.ReadAll(r.Body)
		if err != nil {
			json.NewEncoder(w).Encode(Response{
				Status:  FAILED,
				Message: err.Error(),
			})
			return
		}

		defer r.Body.Close()

		task, err := a.manager.Deploy(dec)
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

func (a *ApiServer) update(w http.ResponseWriter, r *http.Request) {
	a.methodFilter(w, r, []string{"PUT"}, func() {
		dec, err := ioutil.ReadAll(r.Body)
		if err != nil {
			json.NewEncoder(w).Encode(Response{
				Status:  FAILED,
				Message: err.Error(),
			})
			return
		}

		defer r.Body.Close()

		newTask, err := a.manager.Update(dec)
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

func (a *ApiServer) kill(w http.ResponseWriter, r *http.Request) {
	a.methodFilter(w, r, []string{"DELETE"}, func() {
		dec, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}

		defer r.Body.Close()

		err = a.manager.Kill(dec)
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

// Status endpoint lets the end-user know about the TASK_STATUS of their task.
func (a *ApiServer) state(w http.ResponseWriter, r *http.Request) {
	a.methodFilter(w, r, []string{"GET"}, func() {
		name := r.URL.Query().Get("name")
		if name == "" {
			json.NewEncoder(w).Encode(Response{
				Status:  FAILED,
				Message: "No name was found in URL params.",
			})
			return
		}
		state, err := a.manager.Status(name)
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

func (a *ApiServer) tasks(w http.ResponseWriter, r *http.Request) {
	a.methodFilter(w, r, []string{"GET"}, func() {
		tasks, err := a.manager.AllTasks()
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
