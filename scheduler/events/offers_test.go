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
	mockLogger "github.com/verizonlabs/mesos-framework-sdk/logging/test"
	mockResourceManager "github.com/verizonlabs/mesos-framework-sdk/resources/manager/test"
	sched "github.com/verizonlabs/mesos-framework-sdk/scheduler/test"
	"github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"github.com/verizonlabs/mesos-framework-sdk/utils"
	"github.com/verizonlabs/hydrogen/scheduler"
	mockTaskManager "github.com/verizonlabs/hydrogen/task/manager/test"
	mockStorage "github.com/verizonlabs/hydrogen/task/persistence/test"
	"testing"
)

func TestHandler_Offers(t *testing.T) {
	e := NewEventRouter(
		mockTaskManager.MockTaskManager{},
		mockResourceManager.MockResourceManager{},
		new(scheduler.Configuration),
		sched.MockScheduler{},
		&mockStorage.MockStorage{},
		make(chan *manager.Task),
		&mockLogger.MockLogger{},
	)

	// Test empty offers.
	offers := []*mesos_v1.Offer{}
	e.Offers(&mesos_v1_scheduler.Event_Offers{
		Offers: offers,
	})

	// Test a single offer
	offers = []*mesos_v1.Offer{}
	resources := []*mesos_v1.Resource{}
	resources = append(resources, &mesos_v1.Resource{
		Name: utils.ProtoString("cpu"),
		Type: mesos_v1.Value_SCALAR.Enum(),
		Scalar: &mesos_v1.Value_Scalar{
			Value: utils.ProtoFloat64(10.0),
		},
	})

	offers = append(offers, &mesos_v1.Offer{
		Id:          &mesos_v1.OfferID{Value: utils.ProtoString("id")},
		FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString(utils.UuidAsString())},
		AgentId:     &mesos_v1.AgentID{Value: utils.ProtoString(utils.UuidAsString())},
		Hostname:    utils.ProtoString("Some host"),
		Resources:   resources,
	})

	e.Offers(&mesos_v1_scheduler.Event_Offers{
		Offers: offers,
	})
}

func TestHandler_OffersWithQueuedTasks(t *testing.T) {
	e := NewEventRouter(
		mockTaskManager.MockTaskManager{},
		mockResourceManager.MockResourceManager{},
		new(scheduler.Configuration),
		sched.MockScheduler{},
		&mockStorage.MockStorage{},
		make(chan *manager.Task),
		&mockLogger.MockLogger{},
	)

	// Test empty offers.
	offers := []*mesos_v1.Offer{}
	e.Offers(&mesos_v1_scheduler.Event_Offers{
		Offers: offers,
	})

	// Test a single offer
	offers = []*mesos_v1.Offer{}
	resources := []*mesos_v1.Resource{}
	resources = append(resources, &mesos_v1.Resource{
		Name: utils.ProtoString("cpu"),
		Type: mesos_v1.Value_SCALAR.Enum(),
		Scalar: &mesos_v1.Value_Scalar{
			Value: utils.ProtoFloat64(10.0),
		},
	})

	offers = append(offers, &mesos_v1.Offer{
		Id:          &mesos_v1.OfferID{Value: utils.ProtoString("id")},
		FrameworkId: &mesos_v1.FrameworkID{Value: utils.ProtoString(utils.UuidAsString())},
		AgentId:     &mesos_v1.AgentID{Value: utils.ProtoString(utils.UuidAsString())},
		Hostname:    utils.ProtoString("Some host"),
		Resources:   resources,
	})

	e.Offers(&mesos_v1_scheduler.Event_Offers{
		Offers: offers,
	})
}
