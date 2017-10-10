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

package manager

import (
	"github.com/verizonlabs/mesos-framework-sdk/include/mesos_v1"
	"github.com/verizonlabs/mesos-framework-sdk/resources/manager"
	"github.com/verizonlabs/mesos-framework-sdk/task"
	manager2 "github.com/verizonlabs/mesos-framework-sdk/task/manager"
	"github.com/verizonlabs/mesos-framework-sdk/utils"
	"testing"
)

func createResources(cpu, mem float64) (r []*mesos_v1.Resource) {
	r = append(r, &mesos_v1.Resource{
		Name:   utils.ProtoString("cpu"),
		Type:   mesos_v1.Value_SCALAR.Enum(),
		Scalar: &mesos_v1.Value_Scalar{Value: utils.ProtoFloat64(cpu)},
	})

	r = append(r, &mesos_v1.Resource{
		Name:   utils.ProtoString("mem"),
		Type:   mesos_v1.Value_SCALAR.Enum(),
		Scalar: &mesos_v1.Value_Scalar{Value: utils.ProtoFloat64(mem)},
	})
	return
}

func createOffers(num int) (o []*mesos_v1.Offer) {
	for i := 0; i < num; i++ {
		u := utils.UuidAsString()
		o = append(o, &mesos_v1.Offer{
			Id:        &mesos_v1.OfferID{Value: utils.ProtoString(u)},
			Resources: createResources(10.0, 4096.0),
		})
	}
	return
}

func TestNewResourceManager(t *testing.T) {
	rm := manager.NewDefaultResourceManager()
	if rm == nil {
		t.Log("Failed to create a default resource manager.")
		t.FailNow()
	}
}

func TestResourceManager_AddOffers(t *testing.T) {
	rm := manager.NewDefaultResourceManager()
	rm.AddOffers(createOffers(10))
	if length := rm.Offers(); len(length) != 10 {
		t.Logf("Expecting 10, got %v offers", len(length))
	}
}

func TestResourceManager_HasResources(t *testing.T) {
	rm := manager.NewDefaultResourceManager()
	rm.AddOffers(createOffers(1))
	if !rm.HasResources() {
		t.Log("No resources found in the resource manager, expecting some.")
	}
}

func TestResourceManager_Assign(t *testing.T) {
	rm := manager.NewDefaultResourceManager()
	rm.AddOffers(createOffers(10))
	o, err := rm.Assign(
		&manager2.Task{
			Info: &mesos_v1.TaskInfo{
				Name:      utils.ProtoString("test"),
				Resources: createResources(1, 128),
			},
			Filters: []task.Filter{},
		})
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}
	if o == nil {
		t.Log("No offer handed back by Assign.")
		t.FailNow()
	}
	l := rm.Offers()
	if len(l) != 9 {
		t.Logf("Expecting offers to be 9, got %v", len(l))
		t.FailNow()
	}
}

func TestResourceManager_Offers(t *testing.T) {
	rm := manager.NewDefaultResourceManager()
	rm.AddOffers(createOffers(10))
	if o := rm.Offers(); len(o) != 10 {
		t.Logf("Expected 10 offers, got %v", len(o))
		t.FailNow()
	}
}
