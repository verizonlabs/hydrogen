package api

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"mesos-sdk"
	"net/http"
	"sprint/scheduler"
	"sprint/scheduler/server"
	"strconv"
)

const (
	baseUrl        = "/v1/api"
	deployEndpoint = "/deploy"
	statusEndpoint = "/status"
)

//This struct represents the possible application configuration options for an end-user of sprint.
type ApplicationJSON struct {
	Name        string
	Resources   []mesos.Resource
	Command     *mesos.CommandInfo
	Container   *mesos.ContainerInfo
	HealthCheck *mesos.HealthCheck
	Labels      *mesos.Labels
}

type ApiServer struct {
	cfg     server.Configuration
	port    *int
	mux     *http.ServeMux
	handle  map[string]http.HandlerFunc // route -> handler func for that route
	sched   scheduler.SprintScheduler
	version string
}

func NewApiServer(cfg server.Configuration) *ApiServer {
	return &ApiServer{
		cfg:     cfg,
		port:    flag.Int("server.api.port", 8080, "API server listen port"),
		mux:     http.NewServeMux(),
		version: "v1",
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
	a.handle[baseUrl+statusEndpoint] = a.status
}

// RunAPI takes the scheduler controller and sets up the configuration for the API.
func (a *ApiServer) RunAPI() {
	// Set our default handlers here.
	a.setDefaultHandlers()

	// Iterate through all methods and setup endpoints.
	for route, handle := range a.handle {
		a.mux.HandleFunc(route, handle)
	}

	if a.cfg.TLS() {
		a.cfg.Server().Handler = a.mux
		a.cfg.Server().Addr = ":" + strconv.Itoa(*a.port)
		log.Fatal(a.cfg.Server().ListenAndServeTLS(a.cfg.Cert(), a.cfg.Key()))
	} else {
		log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*a.port), a.mux))
	}
}

//Deploy endpoint will parse given JSON and create a given TaskInfo for the scheduler to execute.
func (a *ApiServer) deploy(w http.ResponseWriter, r *http.Request) {
	// We don't want to allow any other methods.
	switch r.Method {
	case "POST":
		{
			dec, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				dec = make([]byte, 0)
				return
			}

			var m ApplicationJSON
			err = json.Unmarshal(dec, &m)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				return
			}

		}
	default:
		{
			fmt.Fprintf(w, r.Method+" is not allowed on this endpoint.")
		}
	}

}

// Status endpoint lets the end-user know about the TASK_STATUS of their task.
func (a *ApiServer) status(w http.ResponseWriter, r *http.Request) {
	// We don't want to allow any other methods.
	switch r.Method {
	case "GET":
		{
			id := r.URL.Query().Get("taskID")

			// Get information about our task status.
			task, err := a.sched.State().TaskSearch(id)
			if err != nil {
				fmt.Fprintf(w, "%v", err.Error())
				return
			}
			fmt.Fprintf(w, "%v", task)

		}
	default:
		{
			fmt.Fprintf(w, r.Method+" is not allowed on this endpoint.")
		}
	}

}
