package manager

import (
	"testing"
	"mesos-framework-sdk/structures"
	"mesos-framework-sdk/logging"
	"sprint/scheduler"
	"mesos-framework-sdk/include/mesos"
	"github.com/gogo/protobuf/proto"
	"strconv"
)

type MockStorage struct {}
func (m *MockStorage) Create(string, ...string) error{
return nil
}
func (m *MockStorage) Read(...string) ([]string, error) {
	return []string{}, nil
}
func (m *MockStorage) Update(string, ...string) error {
	return nil
}
func (m *MockStorage) Delete(string, ...string) error {
	return nil
}
func (m *MockStorage) Driver() string {
	return "mock"
}
func (m *MockStorage) Engine() interface{} {
	return struct {}{}
}

func CreateTestTask(name string) *mesos_v1.TaskInfo {
	return &mesos_v1.TaskInfo{
		Name: proto.String(name),
		TaskId: &mesos_v1.TaskID{Value: proto.String("")},
		AgentId: &mesos_v1.AgentID{Value: proto.String("")},
		Command: &mesos_v1.CommandInfo{
			Value: proto.String("/bin/sleep 50"),
		},
		Container: &mesos_v1.ContainerInfo{
			Type: mesos_v1.ContainerInfo_DOCKER.Enum(),
			Hostname: proto.String("hostname"),
			Mesos: &mesos_v1.ContainerInfo_MesosInfo{
				Image: &mesos_v1.Image{Type: mesos_v1.Image_DOCKER.Enum()},
			},
		},
	}
}


func TestNewTaskManager(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &MockStorage{}
	config := &scheduler.Configuration{}
	logger := logging.NewDefaultLogger()

	taskManager := NewTaskManager(cmap, storage, config, logger)
	if taskManager == nil {
		t.FailNow()
	}
}

func TestTaskManager_Cycle(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &MockStorage{}
	config := &scheduler.Configuration{}
	logger := logging.NewDefaultLogger()
	testTask := CreateTestTask("testTask")

	taskManager := NewTaskManager(cmap, storage, config, logger)
	taskManager.Add(testTask)
	task, err := taskManager.Get(testTask.Name)
	if err != nil {
		t.FailNow()
	}
	if task.String() != testTask.String() {
		t.FailNow()
	}
	taskManager.Delete(task)
	_, err = taskManager.Get(testTask.Name)
	if err == nil {
		t.FailNow()
	}
}

func TestTaskManager_Length(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &MockStorage{}
	config := &scheduler.Configuration{}
	logger := logging.NewDefaultLogger()
	testTask := CreateTestTask("testTask")
	testTask1 := CreateTestTask("testTask1")
	testTask2 := CreateTestTask("testTask2")

	taskManager := NewTaskManager(cmap, storage, config, logger)

	taskManager.Add(testTask)
	taskManager.Add(testTask1)
	taskManager.Add(testTask2)

	task, err := taskManager.Get(testTask.Name)
	if err != nil {
		t.FailNow()
	}
	if task.String() != testTask.String() {
		t.FailNow()
	}
	taskManager.Delete(task)
	_, err = taskManager.Get(testTask.Name)
	if err == nil {
		t.FailNow()
	}

	if taskAmt := taskManager.TotalTasks(); taskAmt != 2 {
		t.Log("Expected 2 total tasks, got " + strconv.Itoa(taskAmt))
		t.FailNow()
	}
	taskManager.Delete(testTask1)
	taskManager.Delete(testTask2)

	if taskAmt := taskManager.TotalTasks(); taskAmt != 0 {
		t.Log("Expected 0 total tasks, got " + strconv.Itoa(taskAmt))
		t.FailNow()
	}
}

func TestTaskManager_GetById(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &MockStorage{}
	config := &scheduler.Configuration{}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := CreateTestTask("testTask")

	taskManager.Add(testTask)

	task, err := taskManager.Get(testTask.Name)
	if err != nil {
		t.Log("GetById failed to get task.")
		t.FailNow()
	}

	taskInfo, err := taskManager.GetById(task.TaskId)
	if err != nil {
		t.Log("GetById got an error.")
		t.FailNow()
	}
	if taskInfo == nil {
		t.Log("GetById got a nil task")
		t.FailNow()
	}
}

func TestTaskManager_GetState(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &MockStorage{}
	config := &scheduler.Configuration{}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := CreateTestTask("testTask")

	taskManager.Add(testTask)

	tasks, err := taskManager.GetState(mesos_v1.TaskState_TASK_UNKNOWN)
	if err != nil {
		t.FailNow()
	}
	if len(tasks) <= 0 || len(tasks) > 1 {
		t.Logf("Tasks returned was %v, expecting 1", len(tasks))
	}
}

func TestTaskManager_HasTask(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &MockStorage{}
	config := &scheduler.Configuration{}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := CreateTestTask("testTask")

	taskManager.Add(testTask)

	if !taskManager.HasTask(testTask){
		t.FailNow()
	}
}

func TestTaskManager_Set(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &MockStorage{}
	config := &scheduler.Configuration{}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := CreateTestTask("testTask")

	taskManager.Add(testTask)

	taskManager.Set(mesos_v1.TaskState_TASK_STAGING, testTask)
	tasks, err := taskManager.GetState(mesos_v1.TaskState_TASK_STAGING)
	if err != nil {
		t.Log("Failed with err on getstate, finished.")
		t.FailNow()
	}
	if len(tasks) <= 0 || len(tasks) > 1 {
		t.Logf("Tasks returned was %v, expecting 1", len(tasks))
	}

	// KILLED and FINISHED delete the task from the task manager.
	taskManager.Set(mesos_v1.TaskState_TASK_KILLED, testTask)

	tasks, err = taskManager.GetState(mesos_v1.TaskState_TASK_KILLED)
	if err == nil {
		t.FailNow()
	}
	if len(tasks) != 0 {
		t.Logf("Tasks returned was %v, expecting 0", len(tasks))
		t.FailNow()
	}
}

func TestTaskManager_TotalTasks(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &MockStorage{}
	config := &scheduler.Configuration{}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := CreateTestTask("testTask")
	testTask1 := CreateTestTask("testTask1")
	testTask2 := CreateTestTask("testTask2")

	taskManager.Add(testTask)
	taskManager.Add(testTask1)
	taskManager.Add(testTask2)

	tasksLength := taskManager.TotalTasks()

	if tasksLength != 3 {
		t.Logf("Expecting 3 tasks, got %v", tasksLength)
		t.FailNow()
	}
	taskManager.Delete(testTask2)

	tasksLength = taskManager.TotalTasks()
	if tasksLength != 2 {
		t.Logf("Expecting 2 tasks, got %v", tasksLength)
		t.FailNow()
	}
	taskManager.Delete(testTask1)

	tasksLength = taskManager.TotalTasks()
	if tasksLength != 1 {
		t.Logf("Expecting 1 tasks, got %v", tasksLength)
		t.FailNow()
	}
}