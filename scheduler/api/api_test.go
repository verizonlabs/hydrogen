package api

import (
	"errors"
	mockLogger "mesos-framework-sdk/logging/test"
	"sprint/scheduler"
	mockApiManager "sprint/scheduler/api/manager/test"
	"testing"
)

type brokenReader struct{}

func (b brokenReader) Read(n []byte) (int, error) {
	return 0, errors.New("I'm broke.")
}

var (
	c      = new(scheduler.Configuration)
	l      = new(mockLogger.MockLogger)
	apiMgr = new(mockApiManager.MockApiManager)
)

// Ensures all components are set correctly when creating the API server.
func TestNewApiServer(t *testing.T) {
	srv := NewApiServer(c, apiMgr, l)
	if srv.cfg != c || srv.manager != apiMgr || srv.logger != l {
		t.Fatal("API does not contain the correct components")
	}
}
