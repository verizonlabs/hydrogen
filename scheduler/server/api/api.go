package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mesos-sdk"
	"net/http"
	"strconv"
)

const (
	baseUrl        = "/v1/api"
	deployEndpoint = "/deploy"
)

/*
This struct represents the possible application configuration options for an end-user of sprint.
*/
type ApplicationJSON struct {
	Name        string
	Resources   []mesos.Resource
	Command     *mesos.CommandInfo
	Container   *mesos.ContainerInfo
	HealthCheck *mesos.HealthCheck
	Labels      *mesos.Labels
}

/*
Configuration for our API
*/
type ApiConfiguration struct {
	port   int
	handle *map[string]http.HandlerFunc // route -> handler func for that route
}

/*
Getter to return our port as an int
*/
func (a *ApiConfiguration) Port() int {
	return a.port
}

/*
Getter to return a string of our port.
*/
func (a *ApiConfiguration) PortAsString() string {
	return strconv.Itoa(a.port)
}

/*
Getter to return our map of handles
*/
func (a *ApiConfiguration) Handle() *map[string]http.HandlerFunc {
	return a.handle
}

/*
Return a default new ApiConfiguration.
*/
func NewApiConfiguration() *ApiConfiguration {
	return &ApiConfiguration{
		handle: defautlHandleFunctions(),
		port:   8080,
	}
}

/*
Set our default API handler routes here.
*/
func defautlHandleFunctions() *map[string]http.HandlerFunc {
	handlers := make(map[string]http.HandlerFunc, 4)
	handlers[baseUrl+deployEndpoint] = deploy
	return &handlers
}

/*
Default Run method with no config
*/
func RunAPI() {
	cfg := NewApiConfiguration()
	// Iterate through all methods and setup endpoints.
	for route, handle := range *cfg.handle {
		http.HandleFunc(route, handle)
	}

	log.Fatal(http.ListenAndServe(":"+cfg.PortAsString(), nil))
}

/*
Run function that takes a given ApiConfiguration to configure the API server.
*/
func Run(cfg *ApiConfiguration) {
	for route, handle := range *cfg.handle {
		http.HandleFunc(route, handle)
	}

	log.Fatal(http.ListenAndServe(":"+cfg.PortAsString(), nil))
}

/*
Deploy endpoint will parse given JSON and create a given TaskInfo for the scheduler to execute.
*/
func deploy(w http.ResponseWriter, r *http.Request) {
	// We don't want to allow any other methods.
	switch r.Method {
	case "POST":
		{
			// Decode the body into slice of bytes, then parse into JSON.
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

			fmt.Fprintf(w, "%v", m) // This is where we would form our Task info.
		}
	default:
		{
			fmt.Fprintf(w, r.Method+" is not allowed on this endpoint.")
		}
	}

}
