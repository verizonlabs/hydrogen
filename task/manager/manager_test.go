package manager

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/task"
	"mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/utils"
	"sprint/scheduler"
	mockStorage "sprint/task/persistence/test"
	"sprint/task/retry"
	"strconv"
	"testing"
	"time"

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
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
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
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
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

func TestTaskManager_State(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := CreateTestTask("testTask")

	taskManager.Add(testTask)

	_, err := taskManager.State(testTask.Name)
	if err != nil {
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
	testTask := CreateTestTask("testTask")

	taskManager.Add(testTask)

	if !taskManager.HasTask(testTask) {
		t.FailNow()
	}
}

func TestTaskManager_Set(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	testTask := CreateTestTask("testTask")

	taskManager.Add(testTask)

	taskManager.Set(mesos_v1.TaskState_TASK_STAGING, testTask)
	_, err := taskManager.State(testTask.Name)
	if err != nil {
		t.Log("Failed with err on getstate, finished.")
		t.FailNow()
	}

	// KILLED and FINISHED delete the task from the task manager.
	taskManager.Set(mesos_v1.TaskState_TASK_FINISHED, testTask)
	_, err = taskManager.State(testTask.Name)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	taskManager.Add(testTask)

	taskManager.Set(mesos_v1.TaskState_TASK_KILLED, testTask)

	_, err = taskManager.State(testTask.Name)
	if err != nil {
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
	testTask := CreateTestTask("testTask")
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
	testTask := CreateTestTask("testTask")
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
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
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
	testTask := CreateTestTask("testTask")
	err := taskManager.Add(testTask)
	if err == nil {
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
	testTask := CreateTestTask("testTask")
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
	testTask := CreateTestTask("testTask")
	taskManager.Set(manager.FAILED, testTask)
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
	tasks := make([]*mesos_v1.TaskInfo, 0)
	for i := 0; i <= 1000; i++ {
		tasks = append(tasks, CreateTestTask("testTask"+strconv.Itoa(i)))
	}
	for _, k := range tasks {
		taskManager.Set(manager.UNKNOWN, k)
	}
	if taskManager.TotalTasks() == 1000 {
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
	tasks := make([]*mesos_v1.TaskInfo, 0)
	for i := 0; i <= 1000; i++ {
		tasks = append(tasks, CreateTestTask("testTask"+strconv.Itoa(i)))
	}
	for _, k := range tasks {
		taskManager.Add(k)
		taskManager.Set(manager.UNKNOWN, k)
	}
	if taskManager.TotalTasks() == 1000 {
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
	tasks := make([]*mesos_v1.TaskInfo, 1000)
	for i := 0; i < 1000; i++ {
		tasks[i] = CreateTestTask("testTask" + strconv.Itoa(i))
	}
	for _, k := range tasks {
		taskManager.Add(k)
		// Try to add the task again, should fail and throw an err.
		if err := taskManager.Add(k); err == nil {
			t.Log("Able to add multiples of the same task", err.Error())
		}
		taskManager.Set(manager.UNKNOWN, k)
	}

	if taskManager.TotalTasks() != 1000 {
		t.Log("Expecting 1000 tasks in total, got " + strconv.Itoa(taskManager.TotalTasks()))
	}
}

func TestSprintTaskHandler_CreateGroup(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)

	if err := taskManager.CreateGroup("Test Group"); err != nil {
		t.Logf("Failed to create a group %v\n", err.Error())
		t.Failed()
	}
}

func TestSprintTaskHandler_AddToGroup(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)

	if err := taskManager.AddToGroup("Test Group", &mesos_v1.AgentID{Value: utils.ProtoString("someAgentId")}); err != nil {
		t.Logf("Failed to add to group %v\n", err.Error())
		t.Failed()
	}
}

func TestSprintTaskHandler_ReadGroup(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)

	if err := taskManager.DelFromGroup("Test Group", &mesos_v1.AgentID{Value: utils.ProtoString("someAgentId")}); err != nil {
		t.Logf("Failed to delete from group %v\n", err.Error())
		t.Failed()
	}
}

func TestSprintTaskHandler_DeleteGroup(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)

	if err := taskManager.DeleteGroup("Test Group"); err != nil {
		t.Logf("Failed to delete group %v\n", err.Error())
		t.Failed()
	}
}

func TestSprintTaskHandler_DelFromGroup(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)

	grouping := taskManager.ReadGroup("Test Group")
	if grouping == nil {
		t.Log("ReadGroup failed.")
		t.Failed()
	}
}

func TestSprintTaskHandler_IsInGroup(t *testing.T) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)

	if ok := taskManager.IsInGroup(&mesos_v1.TaskInfo{Name:utils.ProtoString("Some Task")}); !ok {
		t.Log("Failed to find task in a group")
		t.Failed()
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
	t := CreateTestTask("test")
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
	t := CreateTestTask("test")
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
	t := CreateTestTask("test")
	taskManager.Add(t)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.Get(t.Name)
	}
	b.StopTimer()
}

func BenchmarkSprintTaskHandler_Set(b *testing.B) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	t := CreateTestTask("test")
	taskManager.Add(t)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.Set(manager.UNKNOWN, t)
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
	t := CreateTestTask("test")
	taskManager.Add(t)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.GetById(t.TaskId)
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
	t := CreateTestTask("test")
	taskManager.Add(t)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.HasTask(t)
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
	t := CreateTestTask("test")
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
	t := CreateTestTask("test")
	taskManager.Add(t)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.TotalTasks()
	}
	b.StopTimer()
}

func BenchmarkSprintTaskHandler_State(b *testing.B) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	t := CreateTestTask("test")
	taskManager.Add(t)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.State(t.Name)
	}
	b.StopTimer()
}

func BenchmarkSprintTaskHandler_AddPolicy(b *testing.B) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	t := CreateTestTask("test")
	taskManager.Add(t)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.AddPolicy(&task.TimeRetry{
			Time:       "1",
			Backoff:    true,
			MaxRetries: 100,
		}, t)
	}
	b.StopTimer()
}

func BenchmarkSprintTaskHandler_CheckPolicy(b *testing.B) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	t := CreateTestTask("test")
	taskManager.Add(t)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.CheckPolicy(t)
	}
	b.StopTimer()
}

func BenchmarkSprintTaskHandler_ClearPolicy(b *testing.B) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	t := CreateTestTask("test")
	taskManager.Add(t)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.ClearPolicy(t)
	}
	b.StopTimer()
}

func BenchmarkSprintTaskHandler_RunPolicy(b *testing.B) {
	cmap := make(map[string]manager.Task)
	storage := mockStorage.MockStorage{}
	config := &scheduler.Configuration{
		Persistence: &scheduler.PersistenceConfiguration{
			MaxRetries: 0,
		},
	}
	logger := logging.NewDefaultLogger()
	taskManager := NewTaskManager(cmap, storage, config, logger)
	t := CreateTestTask("test")
	taskManager.Add(t)
	retryFunc := func() error {

		// Check if the task has been deleted while waiting for a retry.
		t, err := taskManager.Get(t.Name)
		if err != nil {
			return err
		}
		taskManager.Set(manager.UNKNOWN, t)

		return nil
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.RunPolicy(&retry.TaskRetry{
			TotalRetries: 0,
			MaxRetries:   100,
			RetryTime:    1 * time.Nanosecond,
			Backoff:      false,
			Name:         "test",
		}, retryFunc)
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
	t := CreateTestTask("test")
	taskManager.Add(t)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		taskManager.AllByState(manager.UNKNOWN)
	}
	b.StopTimer()
}
