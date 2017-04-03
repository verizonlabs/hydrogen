package manager

import (
	"mesos-framework-sdk/resources/manager"
	"testing"
	"mesos-framework-sdk/include/mesos"
	"mesos-framework-sdk/utils"
	"github.com/golang/protobuf/proto"
	"mesos-framework-sdk/task"
)


func createResources(cpu, mem float64) (r []*mesos_v1.Resource){
	r = append(r, &mesos_v1.Resource{
		Name: proto.String("cpu"),
		Type: mesos_v1.Value_SCALAR.Enum(),
		Scalar: &mesos_v1.Value_Scalar{Value: proto.Float64(cpu)},
	})

	r = append(r, &mesos_v1.Resource{
		Name: proto.String("mem"),
		Type: mesos_v1.Value_SCALAR.Enum(),
		Scalar: &mesos_v1.Value_Scalar{Value: proto.Float64(mem)},
	})
	return
}

func createOffers(num int) (o []*mesos_v1.Offer) {
	for i:=0;i < num;i++ {
		u, _ := utils.UuidToString(utils.Uuid())
		o = append(o, &mesos_v1.Offer{
			Id: &mesos_v1.OfferID{Value: proto.String(u)},
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

func TestResourceManager_HasResources(t *testing.T){
	rm := manager.NewDefaultResourceManager()
	rm.AddOffers(createOffers(1))
	if !rm.HasResources() {
		t.Log("No resources found in the resource manager, expecting some.")
	}
}

func TestResourceManager_AddFilter(t *testing.T) {
	rm := manager.NewDefaultResourceManager()
	o := createOffers(10)
	rm.AddOffers(o)
	f := task.Filter{
		Type: "text",
		Value: []string{"test"},
	}
	filters := []task.Filter{f}
	if err := rm.AddFilter(&mesos_v1.TaskInfo{
		Name: proto.String("test"),
		Resources: createResources(1, 128),
	}, filters); err != nil {
		t.Log("Failed to add filter.")
		t.FailNow()
	}
}

func TestResourceManager_Assign(t *testing.T) {
	rm := manager.NewDefaultResourceManager()
	rm.AddOffers(createOffers(10))
	o, err := rm.Assign(&mesos_v1.TaskInfo{
		Name: proto.String("test"),
		Resources: createResources(1, 128),
	})
	if err != nil{
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



