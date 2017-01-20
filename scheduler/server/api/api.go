package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mesos-sdk"
	"net/http"
	"sprint/scheduler"
	"strconv"
	"strings"
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

//Configuration for our API
type ApiConfiguration struct {
	port   int
	handle map[string]http.HandlerFunc // route -> handler func for that route
}

//Getter to return our port as an int
func (a *ApiConfiguration) Port() int {
	return a.port
}

//Getter to return a string of our port.
func (a *ApiConfiguration) PortAsString() string {
	return strconv.Itoa(a.port)
}

//Getter to return our map of handles
func (a *ApiConfiguration) Handle() map[string]http.HandlerFunc {
	return a.handle
}

//Return a default new ApiConfiguration.
func NewApiConfiguration(ctrl *scheduler.SprintScheduler) ApiConfiguration {
	return ApiConfiguration{
		port: 8080,
	}
}

// API struct holds all information regarding the Schedulers API
type API struct {
	version string
	config  ApiConfiguration
	sched   scheduler.SprintScheduler
}

//Set our default API handler routes here.
func (a *API) setDefaultHandlers() {
	a.config.handle = make(map[string]http.HandlerFunc, 4)
	a.config.handle[baseUrl+deployEndpoint] = a.deploy
	a.config.handle[baseUrl+statusEndpoint] = a.status
}

// RunAPI takes the scheduler controller and sets up the configuration for the API.
func RunAPI(schedInstance *scheduler.SprintScheduler) {
	// Set our configuration first.
	api := API{
		version: "v1",
		config:  NewApiConfiguration(schedInstance),
	}
	// Set our default handlers here.
	api.setDefaultHandlers()

	// Iterate through all methods and setup endpoints.
	for route, handle := range api.config.handle {
		http.HandleFunc(route, handle)
	}

	log.Fatal(http.ListenAndServe(":"+api.config.PortAsString(), nil))
}

//Deploy endpoint will parse given JSON and create a given TaskInfo for the scheduler to execute.
func (a *API) deploy(w http.ResponseWriter, r *http.Request) {
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

			log.Printf("%v", m)

		}
	default:
		{
			fmt.Fprintf(w, r.Method+" is not allowed on this endpoint.")
		}
	}

}

// Status endpoint lets the end-user know about the TASK_STATUS of their task.
func (a *API) status(w http.ResponseWriter, r *http.Request) {
	// We don't want to allow any other methods.
	switch r.Method {
	case "GET":
		{
			var id string
			endpoint := fmt.Sprintf("%v", r.URL)

			if strings.Contains(endpoint, "?taskID=") {
				id = strings.SplitAfter(endpoint, "?taskID=")[1]
			}

			// Get information about our task status.
			task, err := a.sched.State().TaskSearch(id)
			if err != nil {
				log.Println(err.Error())
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
