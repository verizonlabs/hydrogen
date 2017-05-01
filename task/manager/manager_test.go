package manager

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/persistence/drivers/etcd/test"
	"mesos-framework-sdk/structures"
	"mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/utils"
	"os"
	"sprint/scheduler"
	"strconv"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	os.Setenv("TESTING", "true")
	ret := m.Run()
	os.Unsetenv("TESTING")
	os.Exit(ret)
}

func CreateTestTask(name string) *mesos_v1.TaskInfo {
	return &mesos_v1.TaskInfo{
		Name:    utils.ProtoString(name),
		TaskId:  &mesos_v1.TaskID{Value: utils.ProtoString("")},
		AgentId: &mesos_v1.AgentID{Value: utils.ProtoString("")},
		Command: &mesos_v1.CommandInfo{
			Value: utils.ProtoString("/bin/sleep 50"),
		},
		Container: &mesos_v1.ContainerInfo{
			Type:     mesos_v1.ContainerInfo_DOCKER.Enum(),
			Hostname: utils.ProtoString("hostname"),
			Mesos: &mesos_v1.ContainerInfo_MesosInfo{
				Image: &mesos_v1.Image{Type: mesos_v1.Image_DOCKER.Enum()},
			},
		},
	}
}

func TestNewTaskManager(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &test.MockKVStore{}
	config := &scheduler.Configuration{}
	logger := logging.NewDefaultLogger()

	taskManager := NewTaskManager(cmap, storage, config, logger)
	if taskManager == nil {
		t.FailNow()
	}
}

func TestTaskManager_Cycle(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &test.MockKVStore{}
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
	storage := &test.MockKVStore{}
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
	storage := &test.MockKVStore{}
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
	storage := &test.MockKVStore{}
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
	storage := &test.MockKVStore{}
	config := &scheduler.Configuration{}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := CreateTestTask("testTask")

	taskManager.Add(testTask)

	if !taskManager.HasTask(testTask) {
		t.FailNow()
	}
}

func TestTaskManager_Set(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &test.MockKVStore{}
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
	taskManager.Set(mesos_v1.TaskState_TASK_FINISHED, testTask)
	tasks, err = taskManager.GetState(mesos_v1.TaskState_TASK_FINISHED)
	if err == nil {
		t.Log(err.Error())
		t.FailNow()
	}
	if len(tasks) != 0 {
		t.Logf("Tasks returned was %v, expecting 0", len(tasks))
	}

	taskManager.Add(testTask)

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
	storage := &test.MockKVStore{}
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

	allTasks := taskManager.Tasks()
	if allTasks.Length() != 1 {
		t.Logf("Expecting 1 tasks, got %v", tasksLength)
		t.FailNow()
	}
}

func TestTaskManager_AddSameTask(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &test.MockKVStore{}
	config := &scheduler.Configuration{}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := CreateTestTask("testTask")
	taskManager.Add(testTask)
	err := taskManager.Add(testTask)
	if err == nil {
		t.Log("Able to add two of the same task, failing.")
		t.FailNow()
	}
}

func TestTaskManager_DeleteFail(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &test.MockBrokenKVStore{}
	config := &scheduler.Configuration{}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := CreateTestTask("testTask")
	taskManager.Add(testTask)
	taskManager.Delete(testTask)
	taskManager.Delete(testTask) // This doesn't work, we need to make sure the storage driver fails.
}

func TestTaskManager_GetByIdFail(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &test.MockKVStore{}
	config := &scheduler.Configuration{}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := CreateTestTask("testTask")
	taskManager.Add(testTask)
	taskManager.Delete(testTask)
	_, err := taskManager.GetById(testTask.GetTaskId())
	if err == nil {
		t.Logf("Found a task by ID after deleting it %v", testTask)
		t.FailNow()
	}

	testTask = CreateTestTask("testTask")
	taskManager.Add(testTask)
	_, err = taskManager.GetById(&mesos_v1.TaskID{Value: utils.ProtoString("Fail me")})
	if err == nil {
		t.Logf("Found a task that never existed: %v", testTask)
		t.FailNow()
	}
}

func TestTaskManager_HasTaskFail(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &test.MockKVStore{}
	config := &scheduler.Configuration{}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := CreateTestTask("testTask")
	taskManager.Add(testTask)
	taskManager.Delete(testTask)
	err := taskManager.HasTask(testTask)
	if err {
		t.Logf("Task manager still thinks it has a task after deleting it %v", testTask)
		t.FailNow()
	}
}

// TODO (tim): This will never pass, logic in manager is to retry forever.
// Need to fix retry logic in manager
func TestTaskManager_HasTaskFailWithBrokenStorage(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &test.MockBrokenKVStore{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			RetryInterval: 1 * time.Second,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := CreateTestTask("testTask")
	err := taskManager.Add(testTask)
	if err == nil {
		t.Log("Didn't fail with broken storage.")
		t.FailNow()
	}
}

func TestTaskManager_DeleteFailWithBrokenStorage(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &test.MockBrokenKVStore{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			RetryInterval: 1 * time.Second,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := CreateTestTask("testTask")
	taskManager.Delete(testTask)
}

func TestTaskManager_SetFailWithBrokenStorage(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &test.MockBrokenKVStore{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			RetryInterval: 1 * time.Second,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := CreateTestTask("testTask")
	taskManager.Set(manager.FAILED, testTask)
}

func TestTaskManager_EncodeFailWithBrokenStorage(t *testing.T) {
	cmap := structures.NewConcurrentMap()
	storage := &test.MockBrokenKVStore{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			RetryInterval: 1 * time.Second,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	taskManager.Add(nil) // Panic will fail testing if it occurs.
}
