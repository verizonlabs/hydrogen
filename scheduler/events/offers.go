// Copyright 2017 Verizon
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package events

import (
	"github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1"
	"github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1_scheduler"
	"github.com/verizonlabs/mesos-framework-sdk/logging"
	"github.com/verizonlabs/mesos-framework-sdk/resources"
	"github.com/verizonlabs/mesos-framework-sdk/scheduler/strategy"
	"github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"github.com/verizonlabs/mesos-framework-sdk/utils"
	"strings"
)

const (
	refuseSeconds = 1.0 // Setting this to 30 as a "reasonable default".
)

//
// Offers is a public method that handles when resource offers are sent to our framework.
// The logic here is based on if we want to accept or decline the offers accordingly.
// If we have no resources or tasks to launch, we simply decline them all,
// otherwise we tell the resource manager to match our tasks with offers sent by the
// master.
//
func (e *Handler) Offers(offerEvent *mesos_v1_scheduler.Event_Offers) {

	// Check if we have any in the task manager we want to launch
	queued, err := e.taskManager.AllByState(manager.UNKNOWN)

	if err != nil {
		e.logger.Emit(logging.INFO, "No tasks to launch")
		_, err := e.scheduler.Suppress()
		if err != nil {
			e.logger.Emit(logging.ERROR, "Failed to suppress offers: %s", err.Error())
		}

		err = e.declineOffers(offerEvent.GetOffers(), refuseSeconds)
		if err != nil {
			e.logger.Emit(logging.ERROR, "Failed to decline offers: %s", err.Error())
		}

		return
	}

	// Update our resources in the manager
	e.resourceManager.AddOffers(offerEvent.GetOffers())
	accepts := make(map[*mesos_v1.OfferID][]*mesos_v1.Offer_Operation)

	for _, task := range queued {
		// If we've hit max retries of a task, kill itself.
		if task.IsKill {
			e.taskManager.Delete(task)
			continue
		}

		if !e.resourceManager.HasResources() {
			break
		}

		offer, err := e.resourceManager.Assign(task)

		if err != nil {
			// It didn't match any offers.
			e.logger.Emit(logging.ERROR, err.Error())
			task.Reschedule(e.revive)
			continue
		}

		if !e.applyStrategy(task, offer) {
			continue
		}
		mesosTask := task.Info
		t := &mesos_v1.TaskInfo{
			Name:        mesosTask.Name,
			TaskId:      mesosTask.GetTaskId(),
			AgentId:     offer.GetAgentId(),
			Command:     mesosTask.GetCommand(),
			Executor:    mesosTask.GetExecutor(),
			Container:   mesosTask.GetContainer(),
			Resources:   mesosTask.GetResources(),
			HealthCheck: mesosTask.GetHealthCheck(),
		}

		if e.config.Executor.CustomExecutor && t.Executor == nil {
			e.setupExecutor(t)
		}

		task.Info = t
		task.State = manager.STAGING

		err = e.taskManager.Update(task)
		if err != nil {
			e.logger.Emit(logging.ERROR, "Failed to update task: %s", err.Error())
		}

		accepts[offer.Id] = append(accepts[offer.Id], resources.LaunchOfferOperation([]*mesos_v1.TaskInfo{t}))
	}

	// Multiplex our tasks onto as few offers as possible and launch them all.
	for id, launches := range accepts {
		// TODO (tim) The offer operations will need to be parsed for volume mounting and etc.)
		_, err := e.scheduler.Accept([]*mesos_v1.OfferID{id}, launches, nil)
		if err != nil {
			e.logger.Emit(logging.ERROR, "Failed to accept offers: %s", err.Error())
		}
	}

	// Resource manager pops offers when they are accepted
	// Offers() returns a list of what is left, therefore whatever is left is to be rejected.
	err = e.declineOffers(e.resourceManager.Offers(), refuseSeconds)
	if err != nil {
		e.logger.Emit(logging.ERROR, "Failed to decline offers: %s", err.Error())
	}
}

func (e *Handler) setupExecutor(t *mesos_v1.TaskInfo) {
	// If we're using our custom executor then make sure we remove the original CommandInfo.
	// Set up our ExecutorInfo and pass the user's command as data to the executor.
	// The executor is responsible for taking this data and acting as expected.
	t.Executor = &mesos_v1.ExecutorInfo{
		ExecutorId: &mesos_v1.ExecutorID{Value: &e.config.Executor.Name},
		Type:       mesos_v1.ExecutorInfo_CUSTOM.Enum(),
		Resources:  t.GetResources(),
		Container:  t.GetContainer(),
		Command:    t.GetCommand(),
		Data:       []byte(t.GetCommand().GetValue()),
	}
	t.Executor.Command.Value = &e.config.Executor.Command
	t.Executor.Command.Uris = []*mesos_v1.CommandInfo_URI{
		{
			Value:      &e.config.Executor.URI,
			Executable: utils.ProtoBool(true),
			Extract:    utils.ProtoBool(false),
			Cache:      utils.ProtoBool(false),
		},
	}
	t.Executor.Command.Shell = &e.config.Executor.Shell
	t.Executor.Command.Arguments = []string{e.config.Executor.Command}

	protocol := "http"
	if e.config.Executor.TLS {
		protocol = "https"
	}

	t.Executor.Command.Environment = &mesos_v1.Environment{Variables: []*mesos_v1.Environment_Variable{
		{
			Name:  utils.ProtoString("PROTOCOL"),
			Value: utils.ProtoString(protocol),
		},
	}}
	t.Command = nil
}

// Decline offers is a private method to organize a list of offers that are to be declined by the
// scheduler.
func (e *Handler) declineOffers(offers []*mesos_v1.Offer, refuseSeconds float64) error {
	if len(offers) == 0 {
		return nil
	}

	declineIDs := make([]*mesos_v1.OfferID, 0, len(offers))

	// Decline whatever offers are left over
	for _, id := range offers {
		declineIDs = append(declineIDs, id.GetId())
	}

	_, err := e.scheduler.Decline(declineIDs, &mesos_v1.Filters{RefuseSeconds: utils.ProtoFloat64(refuseSeconds)})

	return err
}

// Tells us if the strategy the task has is applicable to this offer.
func (e *Handler) applyStrategy(task *manager.Task, offer *mesos_v1.Offer) bool {
	// Apply the strategy.
	switch strings.ToLower(task.Strategy.Type) {
	case strategy.COLOCATE:
	case strategy.UNIQUE:
		if task.GroupInfo.InGroup {
			groupTasks, err := e.taskManager.GetGroup(task)
			if err != nil {
				e.logger.Emit(logging.ERROR, err.Error())
				return false
			}
			for _, groupTask := range groupTasks {
				// If we've already launched onto this agent.
				if groupTask.Info != nil && groupTask.Info.AgentId != nil {
					if groupTask.Info.AgentId.GetValue() == offer.AgentId.GetValue() {
						return false
					}
				}
			}
		} else {
			if t, err := e.taskManager.Get(task.Info.Name); err != nil {
				if t.Info.AgentId != nil {
					if t.Info.AgentId.GetValue() == offer.AgentId.GetValue() {
						return false
					}
				}
			}

		}
	default:
	}
	return true
}
