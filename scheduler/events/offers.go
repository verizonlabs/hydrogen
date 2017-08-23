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
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/resources"
	"mesos-framework-sdk/task/manager"
	"mesos-framework-sdk/utils"
)

//
// Offers is a public method that handles when resource offers are sent to our framework.
// The logic here is based on if we want to accept or decline the offers accordingly.
// If we have no resources or tasks to launch, we simply decline them all,
// otherwise we tell the resource manager to match our tasks with offers sent by the
// master.
//
func (s *SprintEventController) Offers(offerEvent *mesos_v1_scheduler.Event_Offers) {
	// Check if we have any in the task manager we want to launch
	queued, err := s.taskmanager.AllByState(manager.UNKNOWN)

	if err != nil {
		s.logger.Emit(logging.INFO, "No tasks to launch.")
		s.scheduler.Suppress()
		s.declineOffers(offerEvent.GetOffers(), refuseSeconds)
		return
	}

	// Update our resources in the manager
	s.resourcemanager.AddOffers(offerEvent.GetOffers())
	accepts := make(map[*mesos_v1.OfferID][]*mesos_v1.Offer_Operation)

	for _, task := range queued {
		// If we've hit max retries of a task, kill itself.
		if task.IsKill {
			s.taskmanager.Delete(task)
			continue
		}

		if !s.resourcemanager.HasResources() {
			break
		}

		offer, err := s.resourcemanager.Assign(task)

		if err != nil {
			// It didn't match any offers.
			s.logger.Emit(logging.ERROR, err.Error())
			task.Reschedule(s.revive)
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

		if s.config.Executor.CustomExecutor && t.Executor == nil {
			s.setupExecutor(t)
		}

		task.Info = t
		task.State = manager.STAGING

		s.TaskManager().Update(task)

		accepts[offer.Id] = append(accepts[offer.Id], resources.LaunchOfferOperation([]*mesos_v1.TaskInfo{t}))
	}

	// Multiplex our tasks onto as few offers as possible and launch them all.
	for id, launches := range accepts {
		// TODO (tim) The offer operations will need to be parsed for volume mounting and etc.)
		s.scheduler.Accept([]*mesos_v1.OfferID{id}, launches, nil)
	}

	// Resource manager pops offers when they are accepted
	// Offers() returns a list of what is left, therefore whatever is left is to be rejected.
	s.declineOffers(s.ResourceManager().Offers(), refuseSeconds)
}

func (s *SprintEventController) setupExecutor(t *mesos_v1.TaskInfo) {
	// If we're using our custom executor then make sure we remove the original CommandInfo.
	// Set up our ExecutorInfo and pass the user's command as data to the executor.
	// The executor is responsible for taking this data and acting as expected.
	t.Executor = &mesos_v1.ExecutorInfo{
		ExecutorId: &mesos_v1.ExecutorID{Value: &s.config.Executor.Name},
		Type:       mesos_v1.ExecutorInfo_CUSTOM.Enum(),
		Resources:  t.GetResources(),
		Container:  t.GetContainer(),
		Command:    t.GetCommand(),
		Data:       []byte(t.GetCommand().GetValue()),
	}
	t.Executor.Command.Value = &s.config.Executor.Command
	t.Executor.Command.Uris = []*mesos_v1.CommandInfo_URI{
		{
			Value:      &s.config.Executor.URI,
			Executable: utils.ProtoBool(true),
			Extract:    utils.ProtoBool(false),
			Cache:      utils.ProtoBool(false),
		},
	}
	t.Executor.Command.Shell = &s.config.Executor.Shell
	t.Executor.Command.Arguments = []string{s.config.Executor.Command}

	protocol := "http"
	if s.config.Executor.TLS {
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

func (s *SprintEventController) isRedundantOffer(agentId string, agents []*mesos_v1.AgentID) bool {
	for _, i := range agents {
		if agentId == i.GetValue() {
			return true
		}
	}
	return false
}
