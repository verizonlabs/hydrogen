package events

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/include/mesos_v1_scheduler"
	"mesos-framework-sdk/logging"
	"mesos-framework-sdk/resources"
	"mesos-framework-sdk/task/manager"
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
		if !s.resourcemanager.HasResources() {
			break
		}

		offer, err := s.resourcemanager.Assign(mesosTask)
		if err != nil {
			// It didn't match any offers.
			s.logger.Emit(logging.ERROR, err.Error())
			continue // We should decline.
		}

		t := &mesos_v1.TaskInfo{
			Name:      mesosTask.Name,
			TaskId:    mesosTask.GetTaskId(),
			AgentId:   offer.GetAgentId(),
			Command:   mesosTask.GetCommand(),
			Container: mesosTask.GetContainer(),
			Resources: mesosTask.GetResources(),
		}

		// TODO (aaron) investigate this state further as it might cause side effects.
		// this is artificially set to STAGING, it does not correspond to when Mesos sets this task as STAGING.
		// for example other parts of the codebase may check for STAGING and this would cause it to be set too early.
		s.TaskManager().Set(manager.STAGING, t)
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
