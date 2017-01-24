package scheduler

import (
	"flag"
	"io/ioutil"
	"log"
	"mesos-sdk"
	ctrl "mesos-sdk/extras/scheduler/controller"
	"mesos-sdk/httpcli"
	"mesos-sdk/httpcli/httpsched"
	sched "mesos-sdk/scheduler"
	"mesos-sdk/scheduler/calls"
	"os"
	"reflect"
	"testing"
	"time"
)

// Mocked scheduler.
type MockScheduler struct {
	cfg      configuration
	executor *mesos.ExecutorInfo
	state    state
	http     calls.Caller
}

func (m *MockScheduler) NewExecutor() *mesos.ExecutorInfo {
	return &mesos.ExecutorInfo{}
}

func (m *MockScheduler) Config() configuration {
	return m.cfg
}

func (m *MockScheduler) Run(c ctrl.Controller, config *ctrl.Config) error {
	return nil
}

func (m *MockScheduler) State() *state {
	return &m.state
}

func (m *MockScheduler) Caller() *calls.Caller {
	return &m.http
}

func (m *MockScheduler) ExecutorInfo() *mesos.ExecutorInfo {
	return &mesos.ExecutorInfo{}
}

func (m *MockScheduler) FrameworkInfo() *mesos.FrameworkInfo {
	return &mesos.FrameworkInfo{}
}

func (m *MockScheduler) SuppressOffers() error {
	return nil
}

func (m *MockScheduler) ReviveOffers() error {
	return nil
}

func (m *MockScheduler) Reconcile() (mesos.Response, error) {
	return nil, nil
}

// Mocked runner.
// Used for testing the code to run the scheduler.
type mockRunner struct{}

func (m *mockRunner) Run(ctrl.Config) error {
	return nil
}

// Mocked caller.
// Used for testing various things that try and call out to Mesos.
type MockCaller struct{}

func (m *MockCaller) Call(c *sched.Call) (mesos.Response, error) {
	resp := new(mesos.Response)
	return *resp, nil
}

var s Scheduler = &MockScheduler{
	cfg: cfg,
	executor: &mesos.ExecutorInfo{
		ExecutorID: mesos.ExecutorID{
			Value: "",
		},
	},
	state: state{
		totalTasks: 1,
		tasks:      make(map[string]string),
	},
	http: new(MockCaller),
}

// ENTRY POINT FOR ALL TESTS IN THIS PACKAGE
// Suppress our logging and start the tests.
func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	log.SetFlags(0)
	flag.Int("server.executor.port", 0, "test") // Set this for testing purposes
	os.Exit(m.Run())
}

// Ensures we get the correct type back for the scheduler.
func TestNewScheduler(t *testing.T) {
	t.Parallel()

	s := NewScheduler(cfg, make(chan struct{}))

	if reflect.TypeOf(s) != reflect.TypeOf(new(SprintScheduler)) {
		t.Fatal("Controller is not of the right type")
	}
}

// Measures performance of creating a new scheduler.
func BenchmarkNewScheduler(b *testing.B) {
	for n := 0; n < b.N; n++ {
		NewScheduler(cfg, make(chan struct{}))
	}
}

// Make sure we can create new executors correctly.
func TestSprintScheduler_NewExecutor(t *testing.T) {
	t.Parallel()

	s := NewScheduler(cfg, make(chan struct{}))

	executor := s.NewExecutor()
	if executor.GetName() != s.ExecutorInfo().GetName() {
		t.Fatal("Executor name does not match")
	}
	if !reflect.DeepEqual(executor.Command, s.ExecutorInfo().Command) {
		t.Fatal("Executor command does not match")
	}
	if !reflect.DeepEqual(executor.Resources, s.ExecutorInfo().Resources) {
		t.Fatal("Executor resources do not match")
	}
	if executor.GetContainer() != s.ExecutorInfo().GetContainer() {
		t.Fatal("Executor container does not match")
	}
}

// Measures performance of creating new executors.
func BenchmarkSprintScheduler_NewExecutor(b *testing.B) {
	s := NewScheduler(cfg, make(chan struct{}))

	for n := 0; n < b.N; n++ {
		s.NewExecutor()
	}
}

// Checks the configuration stored inside of the scheduler.
func TestSprintScheduler_Config(t *testing.T) {
	t.Parallel()

	s := NewScheduler(cfg, make(chan struct{}))

	cfg := s.Config()
	if reflect.TypeOf(cfg) != reflect.TypeOf(new(SchedulerConfiguration)) {
		t.Fatal("Scheduler configuration is of the wrong type")
	}
	if !*cfg.Checkpointing() {
		t.Fatal("Scheduler checkpointing is not set correctly")
	}
	if *cfg.Command() != "" {
		t.Fatal("Scheduler command is not set properly")
	}
	if cfg.Endpoint() != "http://127.0.0.1:5050/api/v1/scheduler" {
		t.Fatal("Scheduler endpoint is not set correctly")
	}
	if cfg.Name() != "Sprint" {
		t.Fatal("Scheduler name is not set correctly")
	}
	if cfg.MaxRefuse() != 5*time.Second {
		t.Fatal("Scheduler refusal time is not set correctly")
	}
	if cfg.ReviveBurst() != 3 {
		t.Fatal("Scheduler revive burst amount is not set correctly")
	}
	if cfg.ReviveWait() != 1*time.Second {
		t.Fatal("Scheduler revive wait period is not set correctly")
	}
	if cfg.Principal() != "Sprint" {
		t.Fatal("Scheduler principal is not set correctly")
	}
	if cfg.Timeout() != 20*time.Second {
		t.Fatal("Scheduler timeout is not set correctly")
	}
}

// Measures performance of getting the scheduler's configuration.
func BenchmarkSprintScheduler_Config(b *testing.B) {
	s := NewScheduler(cfg, make(chan struct{}))

	for n := 0; n < b.N; n++ {
		s.Config()
	}
}

// Ensures the scheduler's state and contained information is correct.
func TestSprintScheduler_State(t *testing.T) {
	t.Parallel()

	s := NewScheduler(cfg, make(chan struct{}))

	st := s.State()
	if reflect.TypeOf(st) != reflect.TypeOf(new(state)) {
		t.Fatal("Scheduler state is of the wrong type")
	}
	// TODO re-enable this after these are parameterized and hooked up to the API
	/*if st.tasksFinished != 0 || st.tasksLaunched != 0 || st.totalTasks != 0 {
		t.Fatal("Starting state of scheduler tasks is not correct")
	}*/
	if st.done {
		t.Fatal("Scheduler should not be marked as done here")
	}
	if st.frameworkId != "" || st.role != "" {
		t.Fatal("Framework ID and/or role has the wrong starting value")
	}

	token := <-st.reviveTokens
	if token != *new(struct{}) {
		t.Fatal("Couldn't drain revive tokens")
	}

	resource := mesos.Resource{
		Scalar: &mesos.Value_Scalar{
			Value: 64,
		},
		Disk: &mesos.Resource_DiskInfo{
			Volume: &mesos.Volume{
				ContainerPath: "/tmp",
			},
			Source: &mesos.Resource_DiskInfo_Source{
				Mount: &mesos.Resource_DiskInfo_Source_Mount{
					Root: "/tmp",
				},
			},
		},
	}

	st.taskResources.Add(resource)
	if err := st.taskResources.Validate(); err != nil {
		t.Fatal("Failed to validate task resources")
	}
}

// Measures performance of getting the scheduler's state.
func BenchmarkSprintScheduler_State(b *testing.B) {
	s := NewScheduler(cfg, make(chan struct{}))

	for n := 0; n < b.N; n++ {
		s.State()
	}
}

// Tests to see if the scheduler has the right caller.
func TestSprintScheduler_Caller(t *testing.T) {
	t.Parallel()

	s := NewScheduler(cfg, make(chan struct{}))

	caller := s.Caller()
	if reflect.TypeOf(*caller) != reflect.TypeOf(httpsched.NewCaller(httpcli.New())) {
		t.Fatal("Scheduler does not have the right kind of caller")
	}
}

// Measures performance of getting the scheduler's HTTP caller.
func BenchmarkSprintScheduler_Caller(b *testing.B) {
	s := NewScheduler(cfg, make(chan struct{}))

	for n := 0; n < b.N; n++ {
		s.Caller()
	}
}

// Ensures the scheduler's framework info checks out.
func TestSprintScheduler_FrameworkInfo(t *testing.T) {
	t.Parallel()

	s := NewScheduler(cfg, make(chan struct{}))

	info := s.FrameworkInfo()
	if reflect.TypeOf(info) != reflect.TypeOf(new(mesos.FrameworkInfo)) {
		t.Fatal("Scheduler framework info is of the wrong type")
	}
	if !*info.Checkpoint {
		t.Fatal("Checkpointing in scheduler framework info is not set correctly")
	}
	if info.ID != nil {
		t.Fatal("Framework ID should not be set at this point")
	}
	if info.Name != "Sprint" {
		t.Fatal("Framework info contains the wrong framework name")
	}
}

// Measures performance of getting the scheduler's framework information.
func BenchmarkSprintScheduler_FrameworkInfo(b *testing.B) {
	s := NewScheduler(cfg, make(chan struct{}))

	for n := 0; n < b.N; n++ {
		s.FrameworkInfo()
	}
}

// Ensures the scheduler's executor info is valid.
func TestSprintScheduler_ExecutorInfo(t *testing.T) {
	t.Parallel()

	s := NewScheduler(cfg, make(chan struct{}))

	info := s.ExecutorInfo()
	if reflect.TypeOf(info) != reflect.TypeOf(new(mesos.ExecutorInfo)) {
		t.Fatal("Scheduler executor info is of the wrong type")
	}

	// TODO re-enable this once the executor has a decided stable value
	/*id := mesos.ExecutorID{
		Value: "default",
	}
	if info.ExecutorID != id {
		t.Fatal("Executor ID from executor info is incorrect")
	}*/
	if *info.Name != "Sprinter" {
		t.Fatal("Executor has the wrong name")
	}
	// TODO re-enabled this check once more of the executor is parameterized
	/*if info.Command.GetValue() != "" {
		t.Fatal("Executor command has the wrong value")
	}*/
	container := &mesos.ContainerInfo{
		Type: mesos.ContainerInfo_MESOS.Enum(),
	}
	if reflect.TypeOf(info.Container) != reflect.TypeOf(container) {
		t.Fatal("Executor info has the wrong container type")
	}
}

// Measures performance of getting the scheduler's executor information.
func BenchmarkSprintScheduler_ExecutorInfo(b *testing.B) {
	s := NewScheduler(cfg, make(chan struct{}))

	for n := 0; n < b.N; n++ {
		s.ExecutorInfo()
	}
}

// Tests the scheduler's ability to run.
func TestSprintScheduler_Run(t *testing.T) {
	t.Parallel()

	s := NewScheduler(cfg, make(chan struct{}))

	err := s.Run(new(mockRunner), c.BuildConfig(c.BuildContext(), s.Caller(), make(chan struct{}), h))
	if err != nil {
		t.Fatal("Scheduler fails to run properly: " + err.Error())
	}
}

// Measures performance of running the scheduler.
func BenchmarkSprintScheduler_Run(b *testing.B) {
	s := NewScheduler(cfg, make(chan struct{}))

	// Set these up once and only once.
	runner := new(mockRunner)
	config := c.BuildConfig(c.BuildContext(), s.Caller(), make(chan struct{}), h)

	for n := 0; n < b.N; n++ {
		s.Run(runner, config)
	}
}

// Tests offer revival.
func TestSprintScheduler_ReviveOffers(t *testing.T) {
	t.Parallel()

	s := NewScheduler(cfg, make(chan struct{}))
	s.http = new(MockCaller)

	tokens := make(chan struct{}, 1)
	tokens <- struct{}{}
	s.state.reviveTokens = tokens

	if err := s.ReviveOffers(); err != nil {
		t.Fatal("Failed to revive offers: " + err.Error())
	}
}

// Measures performance of reviving offers.
func BenchmarkSprintScheduler_ReviveOffers(b *testing.B) {
	s := NewScheduler(cfg, make(chan struct{}))
	s.http = new(MockCaller)

	tokens := make(chan struct{}, 1)
	s.state.reviveTokens = tokens

	for n := 0; n < b.N; n++ {
		tokens <- struct{}{}
		s.ReviveOffers()
	}
}

// Tests offer suppression.
func TestSprintScheduler_SuppressOffers(t *testing.T) {
	t.Parallel()

	s := NewScheduler(cfg, make(chan struct{}))
	s.http = new(MockCaller)

	if err := s.SuppressOffers(); err != nil {
		t.Fatal("Failed to suppress offers: " + err.Error())
	}
}

// Measures performance of suppressing offers.
func BenchmarkSprintScheduler_SuppressOffers(b *testing.B) {
	s := NewScheduler(cfg, make(chan struct{}))
	s.http = new(MockCaller)

	for n := 0; n < b.N; n++ {
		s.SuppressOffers()
	}
}

// Tests reconciliation.
func TestSprintScheduler_Reconcile(t *testing.T) {
	t.Parallel()

	s := NewScheduler(cfg, make(chan struct{}))
	s.http = new(MockCaller)

	_, err := s.Reconcile()
	if err != nil {
		t.Fatal("Failed to reconcile tasks")
	}
}

// Measures performance of reconciling.
func BenchmarkSprintScheduler_Reconcile(b *testing.B) {
	s := NewScheduler(cfg, make(chan struct{}))
	s.http = new(MockCaller)

	for n := 0; n < b.N; n++ {
		s.Reconcile()
	}
}
