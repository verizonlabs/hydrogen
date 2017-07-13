package manager

import (
	"bytes"
	"encoding/base64"
	"errors"
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/task/manager"
	"strconv"
	"strings"
)

// Private methods that handle operations on tasks.

// Encodes task data to a small, efficient payload that can be transmitted across the wire.
func (m *SprintTaskHandler) encode(task *mesos_v1.TaskInfo, state mesos_v1.TaskState) (bytes.Buffer, error) {
	// Panics on nil values.
	err := m.encoder.Encode(manager.Task{
		Info:  task,
		State: state,
	})
	// NOTE (tim): Since this is inside our struct, we don't need to return it, do we want to change the sig?
	return m.buffer, err
}

// Checks if a task is in a group
func (m *SprintTaskHandler) isInGroup(read Read) {
	_, exists := m.groups[read.name]
	read.isInGroup <- exists
}

// Creates a group for a task.
func (m *SprintTaskHandler) createGroup(write Write) {
	err := m.storage.Create(GROUP_DIRECTORY+write.group, "")
	err = m.storage.Create(GROUP_DIRECTORY+write.group+GROUP_SIZE, strconv.Itoa(0))
	m.groups[write.group] = make([]*mesos_v1.AgentID, 0)
	write.reply <- err
}

// Sets the size of a group.
func (m *SprintTaskHandler) setSize(write Write) {
	// Update size in etcd.
	size, err := m.storage.Read(GROUP_DIRECTORY + write.group + GROUP_SIZE)
	if err != nil {
		write.reply <- err
		return
	}

	// There's not a number stored here, fail.
	s, err := strconv.Atoi(size)
	if err != nil {
		write.reply <- err
		return
	}

	// Increment or decrement by the given size.
	m.storage.Update(GROUP_DIRECTORY+write.group+GROUP_SIZE, strconv.Itoa(s+write.size))

	write.reply <- nil
}

// Read the group name, giving back the agents.
func (m *SprintTaskHandler) readGroup(read Read) {
	read.agents <- m.groups[read.name]
}

// Links an agent to a task group.
func (m *SprintTaskHandler) link(write Write) {
	var newValue string
	currentValues, err := m.storage.Read(GROUP_DIRECTORY + write.group)
	if err != nil {
		write.reply <- err
		return
	}

	if currentValues == "" {
		newValue = write.agentID.GetValue()
	} else {
		newValue = currentValues + "," + write.agentID.GetValue()
	}

	err = m.storage.Update(GROUP_DIRECTORY+write.group, newValue)
	if err != nil {
		write.reply <- err
	}
	write.reply <- nil
}

// Unlinks an agent from a task group.
func (m *SprintTaskHandler) unlink(write Write) {
	// Update values
	currentValues, err := m.storage.Read(GROUP_DIRECTORY + write.group)
	if err != nil {
		write.reply <- err
		return
	}
	split := strings.Split(currentValues, ",")
	var update []string
	for _, i := range split {
		if write.agentID.GetValue() != i {
			update = append(update, i)
		}
	}
	newValue := strings.Join(update, ",")
	m.storage.Update(GROUP_DIRECTORY+write.group, newValue)
	write.reply <- nil
}

// Deletes an entire task group.
func (m *SprintTaskHandler) deleteGroup(write Write) {
	size, err := m.storage.Read(GROUP_DIRECTORY + write.group + GROUP_SIZE)
	if err != nil {
		write.reply <- err
		return
	}
	if s, _ := strconv.Atoi(size); s == 0 {
		err := m.storage.Delete(GROUP_DIRECTORY + write.group)
		if err != nil {
			write.reply <- err
			return
		}
		err = m.storage.Delete(GROUP_DIRECTORY + write.group + GROUP_SIZE)
		if err != nil {
			write.reply <- err
			return
		}
		delete(m.groups, write.group)
	}

	write.reply <- err
}

// Adds a task to the storage backend and data structure.
func (m *SprintTaskHandler) add(add Write) {
	task := add.task

	// Use a unique ID for storing in the map, taskid?
	if _, ok := m.tasks[task.GetName()]; ok {
		add.reply <- errors.New("Task " + task.GetName() + " already exists")
		return
	}

	// Write forward.
	encoded, err := m.encode(task, manager.UNKNOWN)
	if err != nil {
		add.reply <- err
		return
	}

	id := task.TaskId.GetValue()

	// TODO (tim): Policy should be set once for all storage operations and taken care of by storage class.
	// Storage.Create() is called and it's default policy is run.

	policy := m.storage.CheckPolicy(nil)

	err = m.storage.RunPolicy(policy, m.storageWrite(id, encoded))

	if err != nil {
		add.reply <- err
		return
	}

	m.tasks[task.GetName()] = manager.Task{
		State: manager.UNKNOWN,
		Info:  task,
	}

	add.reply <- nil
}

// Deletes a task from the storage backend.
func (m *SprintTaskHandler) delete(res Write) {
	task := res.task
	policy := m.storage.CheckPolicy(nil)
	err := m.storage.RunPolicy(policy, m.storageDelete(task.GetTaskId().GetValue()))

	if err != nil {
		res.reply <- err
		return
	}
	delete(m.tasks, task.GetName())
	m.ClearPolicy(task)

	res.reply <- nil
}

// Gets a task from the data structure.
func (m *SprintTaskHandler) get(res Read) {
	ret := m.tasks[res.name]
	res.reply <- ret.Info
}

// Gets a task by Id from memory.
func (m *SprintTaskHandler) getById(ret Read) {
	id := ret.id
	for _, v := range m.tasks {
		if id == v.Info.GetTaskId().GetValue() {
			ret.reply <- v.Info
			return
		}
	}
	ret.reply <- nil
}

// Checks if a tasks exists.
func (m *SprintTaskHandler) hasTask(ret Read) {
	if val, ok := m.tasks[ret.name]; ok {
		ret.replyTask <- val
	} else {
		ret.replyTask <- manager.Task{}
	}
}

// Returns the total number of tasks known.
func (m *SprintTaskHandler) totalTasks(read Read) {
	read.replySize <- len(m.tasks)
}

// Sets the state of a task.
func (m *SprintTaskHandler) set(ret Write) {
	encoded, err := m.encode(ret.task, ret.state)
	if err != nil {
		ret.reply <- err
		return
	}

	id := ret.task.TaskId.GetValue()

	policy := m.storage.CheckPolicy(nil)
	err = m.storage.RunPolicy(policy, m.storageWrite(id, encoded))

	if err != nil {
		ret.reply <- err
		return
	}

	m.tasks[ret.task.GetName()] = manager.Task{
		Info:  ret.task,
		State: ret.state,
	}
	ret.reply <- nil
}

// Checks the state of a task.
func (m *SprintTaskHandler) state(ret Read) {
	if r, ok := m.tasks[ret.name]; ok {
		ret.replyState <- r.State
	} else {
		ret.replyState <- mesos_v1.TaskState_TASK_UNKNOWN
	}
}

// Gets all tasks that have a particular state.
func (m *SprintTaskHandler) allByState(read Read) {
	for _, v := range m.tasks {
		if v.State == read.state {
			read.reply <- v.Info
		}
	}
	close(read.reply)
}

// Gets all tasks.
func (m *SprintTaskHandler) all(read Read) {
	for _, v := range m.tasks {
		read.replyTask <- v
	}
	close(read.replyTask)
}

// Function that wraps writing to the storage backend.
func (m *SprintTaskHandler) storageWrite(id string, encoded bytes.Buffer) func() error {
	return func() error {
		err := m.storage.Update(TASK_DIRECTORY+id, base64.StdEncoding.EncodeToString(encoded.Bytes()))
		if err != nil {
			m.logger.Emit(
				logging.ERROR, "Failed to update task %s with name %s to persistent data store. Retrying...",
				id,
			)
		}
		return err
	}
}

// Function that wraps deleting from the storage backend.
func (m *SprintTaskHandler) storageDelete(taskId string) func() error {
	return func() error {
		err := m.storage.Delete(TASK_DIRECTORY + taskId)
		if err != nil {
			m.logger.Emit(logging.ERROR, err.Error())
		}
		return err
	}
}
