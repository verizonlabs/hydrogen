package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"strconv"
)

type executorServer struct {
	mux *http.ServeMux
}

// Returns a new instance of our server.
func NewExecutorServer() *executorServer {
	return &executorServer{
		mux: http.NewServeMux(),
	}
}

// Handler to serve the executor binary.
func (s *executorServer) executorHandle(path string, tls bool) {
	s.mux.HandleFunc("/executor", func(w http.ResponseWriter, r *http.Request) {
		if tls {
			// Don't allow fallbacks to HTTP.
			w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		}
		http.ServeFile(w, r, path)
	})
}

// Serve the executor over plain HTTP.
func (s *executorServer) serveExecutor(path string, port int) {
	s.executorHandle(path, false)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), s.mux))
}

// Serve the executor over TLS.
func (s *executorServer) serveExecutorTLS(path string, port int, cert, key string) {
	s.executorHandle(path, true)

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
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

	log.Fatal(srv.ListenAndServeTLS(cert, key))
}

// Start the server
func (s *executorServer) Serve(path string, port int, cert, key string) {
	if cert == "" && key == "" {
		s.serveExecutor(path, port)
	} else {
		s.serveExecutorTLS(path, port, cert, key)
	}
}
