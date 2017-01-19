package api

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	baseUrl        = "/v1/api"
	deployEndpoint = "/deploy"
)

type Message struct {
	Test string
}

type ApiConfiguration struct {
	port   int
	handle map[string]http.HandlerFunc // route -> handler func for that route
}

func RunAPI() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	setupHandle(baseUrl+deployEndpoint, deploy)

	log.Fatal(http.ListenAndServe(":8080", nil))

}

func setupHandle(route string, handle func(w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc(route, handle)
}

// /v1/api/deploy
func deploy(w http.ResponseWriter, r *http.Request) {
	// We don't want to allow any other methods.
	if r.Method == "POST" {
		// Decode the body into slice of bytes, then parse into JSON.
		dec, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}

		var m Message
		err = json.Unmarshal(dec, &m)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}

		fmt.Fprintf(w, "%v", m)
	} else {
		fmt.Fprintf(w, r.Method+" is not allowed on this endpoint.")
	}

}
