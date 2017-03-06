package api

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/server"
	"mesos-framework-sdk/utils"
	"net/http"
	"sprint/scheduler/eventcontroller"
	"strconv"
)

const (
	baseUrl        = "/v1/api"
	deployEndpoint = "/deploy"
	statusEndpoint = "/status"
	killEndpoint   = "/kill"
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
}

type KillJson struct {
	TaskID *string `json:"taskid"`
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
	cfg       server.Configuration
	port      *int
	mux       *http.ServeMux
	handle    map[string]http.HandlerFunc // route -> handler func for that route
	eventCtrl *eventcontroller.SprintEventController
	version   string
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
	a.handle[baseUrl+statusEndpoint] = a.state
	a.handle[baseUrl+killEndpoint] = a.kill
}

// RunAPI takes the scheduler controller and sets up the configuration for the API.
func (a *ApiServer) RunAPI(e *eventcontroller.SprintEventController) {
	// Set our default handlers here.
	a.setDefaultHandlers()

	a.eventCtrl = e

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
				return
			}

			var m ApplicationJSON
			err = json.Unmarshal(dec, &m)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				return
			}

			// Allocate space for our resources.
			var resources []*mesos_v1.Resource
			var scalar = new(mesos_v1.Value_Type)
			*scalar = mesos_v1.Value_SCALAR

			// Add our cpu resources
			var cpu = &mesos_v1.Resource{
				Name:   proto.String("cpu"),
				Type:   scalar,
				Scalar: &mesos_v1.Value_Scalar{Value: proto.Float64(m.Resources.Cpu)},
			}

			// Add our memory resources
			var mem = &mesos_v1.Resource{
				Name:   proto.String("mem"),
				Type:   scalar,
				Scalar: &mesos_v1.Value_Scalar{Value: proto.Float64(m.Resources.Mem)},
			}

			// append into our resources slice.
			resources = append(resources, cpu)
			resources = append(resources, mem)

			uuid, err := utils.UuidToString(utils.Uuid())
			if err != nil {
				log.Println(err.Error())
			}

			task := &mesos_v1.TaskInfo{
				Name:      proto.String(m.Name),
				TaskId:    &mesos_v1.TaskID{Value: proto.String(uuid)},
				Command:   &mesos_v1.CommandInfo{Value: m.Command.Cmd},
				Resources: resources,
				Container: &mesos_v1.ContainerInfo{
					Type: mesos_v1.ContainerInfo_MESOS.Enum(),
					Mesos: &mesos_v1.ContainerInfo_MesosInfo{
						Image: &mesos_v1.Image{
							Docker: &mesos_v1.Image_Docker{
								Name: m.Container.ImageName,
							},
							Type: mesos_v1.Image_DOCKER.Enum(),
						},
					},
					NetworkInfos: []*mesos_v1.NetworkInfo{
						{
							IpAddresses: []*mesos_v1.NetworkInfo_IPAddress{
								{
									Protocol:  mesos_v1.NetworkInfo_IPv6.Enum(),
									IpAddress: proto.String("fe80::3"),
								},
							},
						},
					},
				},
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

func (a *ApiServer) kill(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		{
			// Decode and unmarshal our JSON.
			dec, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return
			}

			var m KillJson
			err = json.Unmarshal(dec, &m)
			if err != nil {
				return
			}

			var status string
			if m.TaskID != nil {
				t := a.eventCtrl.TaskManager().Get(&mesos_v1.TaskID{Value: m.TaskID})
				a.eventCtrl.TaskManager().Delete(t)
				a.eventCtrl.Scheduler().Kill(t.GetTaskId(), t.GetAgentId())
				status = "Task " + t.GetTaskId().GetValue() + " killed."
			} else {
				status = "Task not found"
			}
			fmt.Fprintf(w, "%v", status)
		}
	default:
		{
			fmt.Fprintf(w, r.Method+" is not allowed on this endpoint.")
		}
	}

}

// Status endpoint lets the end-user know about the TASK_STATUS of their task.
func (a *ApiServer) state(w http.ResponseWriter, r *http.Request) {
	// We don't want to allow any other methods.
	switch r.Method {
	case "GET":
		{
			id := r.URL.Query().Get("taskID")

			t := a.eventCtrl.TaskManager().Get(&mesos_v1.TaskID{Value: proto.String(id)})
			queued := a.eventCtrl.TaskManager().QueuedTasks()
			var status string
			if _, ok := queued[t.GetTaskId().GetValue()]; ok {
				status = "task " + t.GetTaskId().GetValue() + " is queued."
			} else {
				status = "task " + t.GetTaskId().GetValue() + " is launched."
			}
			fmt.Fprintf(w, "%v", status)

		}
	default:
		{
			fmt.Fprintf(w, r.Method+" is not allowed on this endpoint.")
		}
	}

}
