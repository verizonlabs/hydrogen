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

	for _, mesosTask := range queued {
		var InGroup bool
		if !s.resourcemanager.HasResources() {
			break
		}

		//
		// Check if the task is in a scale group
		// scale group will hold relevant information
		// about which tasks are on which agents.
		//

		group := s.taskmanager.ReadGroup(mesosTask.GetName())

		if group != nil {
			InGroup = true
		}

		offer, err := s.resourcemanager.Assign(mesosTask)

		// If we're in a group and we've already launched onto this agent, skip it.
		if InGroup && s.isRedundantOffer(offer.GetAgentId().GetValue(), group) {
			continue
		}

		if err != nil {
			// It didn't match any offers.
			s.logger.Emit(logging.ERROR, err.Error())
			continue // We should decline.
		}

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

		// If we're using our custom executor then make sure we remove the original CommandInfo.
		// Set up our ExecutorInfo and pass the user's command as data to the executor.
		// The executor is responsible for taking this data and acting as expected.
		if s.config.Executor.CustomExecutor && t.Executor == nil {
			t.Executor = &mesos_v1.ExecutorInfo{
				ExecutorId: &mesos_v1.ExecutorID{Value: &s.config.Executor.Name},
				Type:       mesos_v1.ExecutorInfo_CUSTOM.Enum(),
				Resources:  mesosTask.GetResources(),
				Container:  mesosTask.GetContainer(),
				Command:    mesosTask.GetCommand(),
				Data:       []byte(mesosTask.GetCommand().GetValue()),
			}
			// This executor command should simply be ./executor
			t.Executor.Command.Value = &s.config.Executor.Command
			t.Executor.Command.Uris = []*mesos_v1.CommandInfo_URI{
				{
					Value:      &s.config.FileServer.Path,
					Executable: utils.ProtoBool(true),
					Extract:    utils.ProtoBool(false),
					Cache:      utils.ProtoBool(false),
				},
			}
			// TODO (tim): change the default to true for now until we have support for shell-less executors.
			t.Executor.Command.Shell = &s.config.Executor.Shell
			// NOTE: Tell users in README about how custom executors work? How do we support users with custom
			// commands with an executor? Can run executor && <user_command>...
			t.Command = nil
		}
		// TODO (aaron) investigate this state further as it might cause side effects.
		// this is artificially set to STAGING, it does not correspond to when Mesos sets this task as STAGING.
		// for example other parts of the codebase may check for STAGING and this would cause it to be set too early.
		s.TaskManager().Set(manager.STAGING, t)
		if InGroup {
			s.TaskManager().Link(mesosTask.GetName(), offer.GetAgentId())
		}
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

func (s *SprintEventController) isRedundantOffer(agentId string, agents []*mesos_v1.AgentID) bool {
	for _, i := range agents {
		if agentId == i.GetValue() {
			return true
		}
	}
	return false
}
