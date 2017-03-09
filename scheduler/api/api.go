package api

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/logging"
	resourcebuilder "mesos-framework-sdk/resources"
	"mesos-framework-sdk/server"
	taskbuilder "mesos-framework-sdk/task"
	"mesos-framework-sdk/utils"
	"net/http"
	"sprint/scheduler/events"
	"strconv"
	"time"
)

const (
	baseUrl        = "/v1/api"
	deployEndpoint = "/deploy"
	statusEndpoint = "/status"
	killEndpoint   = "/kill"
	updateEndpoint = "/update"
	retries        = 20
)

type ApiServer struct {
	cfg       server.Configuration
	port      *int
	mux       *http.ServeMux
	handle    map[string]http.HandlerFunc // route -> handler func for that route
	eventCtrl *eventcontroller.SprintEventController
	version   string
	log       *logging.DefaultLogger
}

func NewApiServer(cfg server.Configuration, mux *http.ServeMux, port *int, version string, lgr *logging.DefaultLogger) *ApiServer {
	return &ApiServer{
		cfg:     cfg,
		port:    port,
		mux:     mux,
		version: version,
		log:     lgr,
	}
}

//Getter to return a map of handlers
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
}

func (a *ApiServer) setHandlers(handles map[string]http.HandlerFunc) {
	for route, handle := range handles {
		a.handle[route] = handle
	}
}

func (a *ApiServer) setEventController(e *eventcontroller.SprintEventController) {
	a.eventCtrl = e
}

// RunAPI takes the scheduler controller and sets up the configuration for the API.
func (a *ApiServer) RunAPI(e *eventcontroller.SprintEventController, handlers map[string]http.HandlerFunc) {
	if handlers != nil || len(handlers) == 0 {
		a.setHandlers(handlers)
	} else {
		a.log.Emit(logging.INFO, "Setting default handlers instead since handlers passed in was nil or empty.")
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
		a.log.Emit(logging.ERROR, a.cfg.Server().ListenAndServeTLS(a.cfg.Cert(), a.cfg.Key()).Error())
	} else {
		a.log.Emit(logging.ERROR, http.ListenAndServe(":"+strconv.Itoa(*a.port), a.mux).Error())
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

			// Allocate space for our resources.
			var resources []*mesos_v1.Resource

			var cpu = resourcebuilder.CreateCpu(m.Resources.Cpu, "*")
			var mem = resourcebuilder.CreateMem(m.Resources.Mem, "*")

			networks, err := taskbuilder.ParseNetworkJSON(m.Container.Network)
			if err != nil {
				// This isn't a fatal error so we can log this as debug and move along.
				a.log.Emit(logging.INFO, "No explicit network info passed in.")
			}

			// append into our resources slice.
			resources = append(resources, cpu, mem)

			uuid, err := utils.UuidToString(utils.Uuid())
			if err != nil {
				a.log.Emit(logging.ERROR, err.Error())
			}

			task := &mesos_v1.TaskInfo{
				Name:      proto.String(m.Name),
				TaskId:    &mesos_v1.TaskID{Value: proto.String(uuid)},
				Command:   resourcebuilder.CreateSimpleCommandInfo(m.Command.Cmd, nil),
				Resources: resources,
				Container: &mesos_v1.ContainerInfo{
					Type: mesos_v1.ContainerInfo_MESOS.Enum(),
					Mesos: &mesos_v1.ContainerInfo_MesosInfo{
						Image: &mesos_v1.Image{
							Type: mesos_v1.Image_DOCKER.Enum(),
							Docker: &mesos_v1.Image_Docker{
								Name: m.Container.ImageName,
							},
						},
					},
					NetworkInfos: networks,
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

			var resources []*mesos_v1.Resource

			var cpu = resourcebuilder.CreateCpu(m.Resources.Cpu, "*")
			var mem = resourcebuilder.CreateMem(m.Resources.Mem, "*")

			networks, err := taskbuilder.ParseNetworkJSON(m.Container.Network)
			if err != nil {
				// This isn't a fatal error so we can log this as debug and move along.
				a.log.Emit(logging.INFO, "No explicit network info passed in.")
			}

			// append into our resources slice.
			resources = append(resources, cpu, mem)

			uuid, err := utils.UuidToString(utils.Uuid())
			if err != nil {
				a.log.Emit(logging.ERROR, err.Error())
			}

			task := &mesos_v1.TaskInfo{
				Name:      proto.String(m.Name),
				TaskId:    &mesos_v1.TaskID{Value: proto.String(uuid)},
				Command:   resourcebuilder.CreateSimpleCommandInfo(m.Command.Cmd, nil),
				Resources: resources,
				Container: &mesos_v1.ContainerInfo{
					Type: mesos_v1.ContainerInfo_MESOS.Enum(),
					Mesos: &mesos_v1.ContainerInfo_MesosInfo{
						Image: &mesos_v1.Image{
							Type: mesos_v1.Image_DOCKER.Enum(),
							Docker: &mesos_v1.Image_Docker{
								Name: m.Container.ImageName,
							},
						},
					},
					NetworkInfos: networks,
				},
			}
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
					a.eventCtrl.TaskManager().Delete(t)
					a.eventCtrl.Scheduler().Kill(t.GetTaskId(), t.GetAgentId())
					status = "Task " + t.GetName() + " killed."
				} else {
					status = "Unable to retrieve task from task manager."
				}
			} else {
				status = "Task not found."
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
			if _, ok := queued[t.GetTaskId().GetValue()]; ok {
				status = "task " + t.GetName() + " is queued."
			} else {
				status = "task " + t.GetName() + " is launched."
			}
			fmt.Fprintf(w, "%v", status)
		}
	default:
		{
			fmt.Fprintf(w, r.Method+" is not allowed on this endpoint.")
		}
	}

}
