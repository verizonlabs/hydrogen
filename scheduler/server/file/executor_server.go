package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"strconv"
)

type executorServer struct {
	mux  *http.ServeMux
	path string
	port int
	tls  bool
	cert string
	key  string
}

// Returns a new instance of our server.
func NewExecutorServer(cfg Configuration) *executorServer {
	return &executorServer{
		mux:  http.NewServeMux(),
		path: cfg.ExecutorSrvPath(),
		port: cfg.ExecutorSrvPort(),
		tls:  cfg.ExecutorSrvCert() != "" && cfg.ExecutorSrvKey() != "",
		cert: cfg.ExecutorSrvCert(),
		key:  cfg.ExecutorSrvKey(),
	}
}

// Maps endpoints to handlers.
func (s *executorServer) executorHandlers(path string) {
	s.mux.HandleFunc("/executor", s.executorBinary)
}

// Serve the executor binary.
func (s *executorServer) executorBinary(w http.ResponseWriter, r *http.Request) {
	_, err := os.Stat(s.path) // check if the file exists first.
	if err != nil {
		log.Fatal(s.path + " does not exist. " + err.Error())
	}

	if s.tls {
		// Don't allow fallbacks to HTTP.
		w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	}
	http.ServeFile(w, r, s.path)
}

// Start the server with or without TLS.
func (s *executorServer) Serve() {
	s.executorHandlers(s.path)

	if s.tls {
		srv := &http.Server{
			Addr:    ":" + strconv.Itoa(s.port),
			Handler: s.mux,
			TLSConfig: &tls.Config{
				// Use only the most secure protocol version.
				MinVersion: tls.VersionTLS12,
				// Use very strong crypto curves.
				CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
				PreferServerCipherSuites: true,
				// Use very strong cipher suites (order is important here!)
				CipherSuites: []uint16{
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, // Required for HTTP/2 support.
					tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
					tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				},
			},
		}

		log.Fatal(srv.ListenAndServeTLS(s.cert, s.key))
	} else {
		log.Fatal(http.ListenAndServe(":"+strconv.Itoa(s.port), s.mux))
	}
}
