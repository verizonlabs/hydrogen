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
//Standardized to lower case.
type ApplicationJSON struct {
	Name        string              `json:"name"`
	Resources   *ResourceJSON       `json:"resources"`
	Command     *CommandJSON        `json:"command"`
	Container   *ContainerJSON      `json:"container"`
	HealthCheck *HealthCheckJSON    `json:"healthcheck"`
	Labels      []map[string]string `json:"labels"`
}

// How do we want to define health checks?
// Scripts, api end points, timers...etc?
type HealthCheckJSON struct {
	Endpoint *string `json:"endpoint"`
	// ?
}

//Struct to define our resources
type ResourceJSON struct {
	Mem  float64 `json:"mem"`
	Cpu  float64 `json:"cpu"`
	Disk float64 `json:"disk"`
}

//Struct to define a command for our container.
type CommandJSON struct {
	Cmd  *string   `json:"cmd"`
	Uris []UriJSON `json:"uris"`
}

//Struct to define our container image and tag.
type ContainerJSON struct {
	ImageName *string `json:"image"`
	Tag       *string `json:"tag"`
}

//Struct to define our URI resources
type UriJSON struct {
	Uri     *string `json:"uri"`
	Extract *bool   `json:"extract"`
	Execute *bool   `json:"execute"`
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
			// Decode and unmarshal our JSON.
			dec, err := ioutil.ReadAll(r.Body)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				dec = make([]byte, 0) // is this necessary?
				return
			}

			var m ApplicationJSON
			err = json.Unmarshal(dec, &m)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				return
			}

			// Allocate space for our resources.
			var resources []mesos.Resource
			var scalar = new(mesos.Value_Type)
			*scalar = mesos.SCALAR

			// Add our cpu resources
			var cpu = mesos.Resource{
				Name:   "cpu",
				Type:   scalar,
				Scalar: &mesos.Value_Scalar{Value: m.Resources.Cpu},
			}

			// Add our memory resources
			var mem = mesos.Resource{
				Name:   "mem",
				Type:   scalar,
				Scalar: &mesos.Value_Scalar{Value: m.Resources.Mem},
			}

			// append into our resources slice.
			resources = append(resources, cpu)
			resources = append(resources, mem)

			// TODO: diskinfo resources + external disks.

			var command = &mesos.CommandInfo{
				Value: m.Command.Cmd,
			}

			// Setup our docker container from the API call
			var docker = new(mesos.ContainerInfo_Type)
			*docker = mesos.ContainerInfo_DOCKER
			var container = &mesos.ContainerInfo{
				Type: docker,
				Docker: &mesos.ContainerInfo_DockerInfo{
					Image: *m.Container.ImageName + ":" + *m.Container.Tag,
				},
			}

			// Final constructed task info
			var taskInfo = mesos.TaskInfo{
				Name:      m.Name,
				TaskID:    mesos.TaskID{},
				Resources: resources,
				Command:   command,
				Container: container,
			}
			a.sched.State().AddTask(taskInfo) // Update our scheduler with the new task.
			fmt.Fprintf(w, "%v", taskInfo.String())
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
