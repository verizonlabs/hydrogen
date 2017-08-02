package manager

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/utils"
	"sprint/scheduler"
	mockStorage "sprint/task/persistence/test"
	"strconv"
	"testing"
)

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
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()

	taskManager := NewTaskManager(cmap, storage, config, logger)
	if taskManager == nil {
		t.FailNow()
	}
}

func TestTaskManager_Cycle(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	testTask := CreateTestTask("testTask")

	taskManager := NewTaskManager(cmap, storage, config, logger)
	taskManager.Add(&manager.Task{Info: testTask, Instances: 1, State: manager.UNKNOWN})
	task, err := taskManager.Get(testTask.Name)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	if task.Info.String() != testTask.String() {
		t.Log(err)
		t.FailNow()
	}
	taskManager.Delete(task)
	_, err = taskManager.Get(testTask.Name)
	if err == nil {
		t.Log(err)
		t.FailNow()
	}
}

func TestTaskManager_Length(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	testTask := &manager.Task{Info: CreateTestTask("testTask"), Instances: 1, State: manager.UNKNOWN}
	testTask1 := &manager.Task{Info: CreateTestTask("testTask1"), Instances: 1, State: manager.UNKNOWN}
	testTask2 := &manager.Task{Info: CreateTestTask("testTask2"), Instances: 1, State: manager.UNKNOWN}

	taskManager := NewTaskManager(cmap, storage, config, logger)

	taskManager.Add(testTask)
	taskManager.Add(testTask1)
	taskManager.Add(testTask2)

	tsk, err := taskManager.Get(testTask.Info.Name)
	if err != nil {
		t.FailNow()
	}
	if tsk.Info.String() != testTask.Info.String() {
		t.FailNow()
	}
	taskManager.Delete(tsk)
	_, err = taskManager.Get(tsk.Info.Name)
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
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := &manager.Task{Info: CreateTestTask("testTask"), Instances: 1, State: manager.UNKNOWN}

	taskManager.Add(testTask)

	task, err := taskManager.Get(testTask.Info.Name)
	if err != nil {
		t.Log("GetById failed to get task.")
		t.FailNow()
	}

	taskInfo, err := taskManager.GetById(task.Info.TaskId)
	if err != nil {
		t.Log("GetById got an error.")
		t.FailNow()
	}
	if taskInfo == nil {
		t.Log("GetById got a nil task")
		t.FailNow()
	}
}

func TestTaskManager_HasTask(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := &manager.Task{Info: CreateTestTask("testTask"), Instances: 1, State: manager.UNKNOWN}

	taskManager.Add(testTask)

	if !taskManager.HasTask(testTask.Info) {
		t.FailNow()
	}
}

func TestTaskManager_TotalTasks(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := &manager.Task{Info: CreateTestTask("testTask"), Instances: 1, State: manager.UNKNOWN}
	testTask1 := &manager.Task{Info: CreateTestTask("testTask1"), Instances: 1, State: manager.UNKNOWN}
	testTask2 := &manager.Task{Info: CreateTestTask("testTask2"), Instances: 1, State: manager.UNKNOWN}

	taskManager.Add(testTask, testTask1, testTask2)

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

func TestTaskManager_AddSameTask(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := &manager.Task{Info: CreateTestTask("testTask"), Instances: 1, State: manager.UNKNOWN}
	taskManager.Add(testTask)
	err := taskManager.Add(testTask)
	if err == nil {
		t.Log("Able to add two of the same task, failing.")
		t.FailNow()
	}
}

func TestTaskManager_DeleteFail(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := &manager.Task{Info: CreateTestTask("testTask"), State: manager.UNKNOWN}
	taskManager.Add(testTask)
	taskManager.Delete(testTask)
	taskManager.Delete(testTask) // This doesn't work, we need to make sure the storage driver fails.
}

func TestTaskManager_GetByIdFail(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := &manager.Task{Info: CreateTestTask("testTask"), State: manager.UNKNOWN}
	taskManager.Add(testTask)
	taskManager.Delete(testTask)
	_, err := taskManager.GetById(testTask.Info.GetTaskId())
	if err == nil {
		t.Logf("Found a task by ID after deleting it %v", testTask)
		t.FailNow()
	}

	testTask = &manager.Task{Info: CreateTestTask("testTask"), State: manager.UNKNOWN}
	taskManager.Add(testTask)
	_, err = taskManager.GetById(&mesos_v1.TaskID{Value: utils.ProtoString("Fail me")})
	if err == nil {
		t.Logf("Found a task that never existed: %v", testTask)
		t.FailNow()
	}
}

func TestTaskManager_HasTaskFail(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := &manager.Task{Info: CreateTestTask("testTask"), State: manager.UNKNOWN}
	taskManager.Add(testTask)
	taskManager.Delete(testTask)
	err := taskManager.HasTask(testTask.Info)
	if err {
		t.Logf("Task manager still thinks it has a task after deleting it %v", testTask)
		t.FailNow()
	}
}

func TestTaskManager_HasTaskFailWithBrokenStorage(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockBrokenStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := &manager.Task{Info: CreateTestTask("testTask"), Instances: 1, State: manager.UNKNOWN}
	err := taskManager.Add(testTask)
	if err == nil {
		t.Log(err)
		t.Log("Didn't fail with broken storage.")
		t.FailNow()
	}
}

func TestTaskManager_DeleteFailWithBrokenStorage(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := &manager.Task{Info: CreateTestTask("testTask"), State: manager.UNKNOWN}
	taskManager.Delete(testTask)
}

func TestTaskManager_SetFailWithBrokenStorage(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := &manager.Task{Info: CreateTestTask("testTask"), State: manager.UNKNOWN}
	taskManager.Add(testTask)
}

func TestTaskManager_AddManyTasks(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	tasks := make([]*manager.Task, 0)
	for i := 0; i <= 3; i++ {
		tasks = append(tasks, &manager.Task{Info: CreateTestTask("testTask"), State: manager.UNKNOWN})
	}
	taskManager.Add(tasks...)
	if taskManager.TotalTasks() == 3 {
		t.Log(taskManager.TotalTasks())
	}
}

func TestTaskManager_AddManyTasksAndDelete(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	tasks := make([]*manager.Task, 0)
	for i := 0; i <= 3; i++ {
		tasks = append(tasks, &manager.Task{Info: CreateTestTask("testTask"), State: manager.UNKNOWN})
	}
	for _, k := range tasks {
		taskManager.Add(k)
	}
	if taskManager.TotalTasks() == 3 {
		t.Log(taskManager.TotalTasks())
	}
	for _, k := range tasks {
		taskManager.Delete(k)
	}
}

func TestTaskManager_DoubleAdd(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	tasks := []*manager.Task{
		{Info: CreateTestTask("testTask"), Instances: 1, State: manager.UNKNOWN},
		{Info: CreateTestTask("testTask"), Instances: 1, State: manager.UNKNOWN},
	}
	// This should fail since it's two of the same tasks.
	if err := taskManager.Add(tasks...); err == nil {
		t.Log("Able to add multiples of the same task", err.Error())
	}
}

func BenchmarkSprintTaskHandler_Add(b *testing.B) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	t := &manager.Task{Info: CreateTestTask("testTask"), State: manager.UNKNOWN}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.Add(t)
	}
	b.StopTimer()
}

func BenchmarkSprintTaskHandler_Delete(b *testing.B) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	t := &manager.Task{Info: CreateTestTask("testTask"), State: manager.UNKNOWN}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.Delete(t)
	}
	b.StopTimer()

}

func BenchmarkSprintTaskHandler_Get(b *testing.B) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	t := &manager.Task{Info: CreateTestTask("testTask"), State: manager.UNKNOWN}
	taskManager.Add(t)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.Get(t.Info.Name)
	}
	b.StopTimer()
}

func BenchmarkSprintTaskHandler_GetById(b *testing.B) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	t := &manager.Task{Info: CreateTestTask("testTask"), State: manager.UNKNOWN}
	taskManager.Add(t)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.GetById(t.Info.TaskId)
	}
	b.StopTimer()
}

func BenchmarkSprintTaskHandler_HasTask(b *testing.B) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	t := &manager.Task{Info: CreateTestTask("testTask"), State: manager.UNKNOWN}
	taskManager.Add(t)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.HasTask(t.Info)
	}
	b.StopTimer()
}

func BenchmarkSprintTaskHandler_All(b *testing.B) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	t := &manager.Task{Info: CreateTestTask("testTask"), State: manager.UNKNOWN}
	taskManager.Add(t)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.All()
	}
	b.StopTimer()
}

func BenchmarkSprintTaskHandler_TotalTasks(b *testing.B) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	t := &manager.Task{Info: CreateTestTask("testTask"), State: manager.UNKNOWN}
	taskManager.Add(t)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.TotalTasks()
	}
	b.StopTimer()
}

func BenchmarkSprintTaskHandler_AllByState(b *testing.B) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	t := &manager.Task{Info: CreateTestTask("testTask"), State: manager.UNKNOWN}
	taskManager.Add(t)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.AllByState(manager.UNKNOWN)
	}
	b.StopTimer()
}

func BenchmarkSprintTaskHandler_Update(b *testing.B) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	t := &manager.Task{Info: CreateTestTask("testTask"), State: manager.UNKNOWN}
	taskManager.Add(t)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.Update(t)
	}
	b.StopTimer()
}
