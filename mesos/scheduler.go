package mesos

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	"github.com/mesos/mesos-go/mesosutil"
	sched "github.com/mesos/mesos-go/scheduler"
)

type Task struct {
	Id        string
	Name      string
	scheduled bool
}

func (t *Task) Schedule() {
	t.scheduled = true
}

func (t *Task) IsScheduled() bool {
	return t.scheduled
}

type rancherScheduler struct {
	taskChan        chan *Task
	rancherExecutor *mesos.ExecutorInfo
}

const (
	taskCPUs = 0.1
	taskMem  = 32.0
)

var (
	defaultFilter = &mesos.Filters{RefuseSeconds: proto.Float64(1)}
)

func newRancherMesosScheduler(taskChan chan *Task) *rancherScheduler {
	command := "rancher-mesos-executor --work_dir=/home/wlan0/Media/hdd"
	return &rancherScheduler{
		taskChan: taskChan,
		rancherExecutor: &mesos.ExecutorInfo{
			ExecutorId: &mesos.ExecutorID{Value: proto.String("rancher-executor")},
			Command: &mesos.CommandInfo{
				Value: proto.String(command),
				Uris:  []*mesos.CommandInfo_URI{},
			},
			Name: proto.String("RancherExecutor"),
		},
	}
}

func (s *rancherScheduler) newRancherTask(task *Task, offer *mesos.Offer) *mesos.TaskInfo {
	data := map[string]string{}
	data["cattle_url"] = "http://192.168.11.212:8080"
	data["agent_version"] = "v0.7.11"
	data["registration_token"] = "AC4C49FEF7D2D162DBD9:1438128000000:UEeVaQvHRB8iopQlIZQ9VoiXro"
	dataBytes, _ := json.Marshal(data)
	t := &mesos.TaskInfo{
		TaskId: &mesos.TaskID{
			Value: proto.String(fmt.Sprintf("RANCHER-%d", task.Id)),
		},
		SlaveId: offer.SlaveId,
		Resources: []*mesos.Resource{
			mesosutil.NewScalarResource("cpus", taskCPUs),
			mesosutil.NewScalarResource("mem", taskMem),
		},
		Data: dataBytes,
	}
	name := fmt.Sprintf("Rancher-%d", task.Id)
	t.Name = &name
	t.Executor = s.rancherExecutor
	return t
}

func (s *rancherScheduler) Registered(
	_ sched.SchedulerDriver,
	frameworkID *mesos.FrameworkID,
	masterInfo *mesos.MasterInfo) {
	fmt.Println("Registered")
}

func (s *rancherScheduler) Reregistered(_ sched.SchedulerDriver, masterInfo *mesos.MasterInfo) {
	fmt.Println("Registered")
}

func (s *rancherScheduler) Disconnected(sched.SchedulerDriver) {
	fmt.Println("disconnected")
}

func (s *rancherScheduler) ResourceOffers(driver sched.SchedulerDriver, offers []*mesos.Offer) {
	fmt.Println("Got a resource offer...")
	for _, offer := range offers {

		tasks := []*mesos.TaskInfo{}
		select {
		case res := <-s.taskChan:
			tasks = append(tasks, s.newRancherTask(res, offer))
		case <-time.After(10 * time.Second):
			driver.DeclineOffer(offer.Id, defaultFilter)
			continue
		}
		driver.LaunchTasks([]*mesos.OfferID{offer.Id}, tasks, defaultFilter)
	}
}

func (s *rancherScheduler) Error(_ sched.SchedulerDriver, err string) {
	fmt.Println("error", err)
}

func (s *rancherScheduler) StatusUpdate(driver sched.SchedulerDriver, status *mesos.TaskStatus) {
	fmt.Println("Status update")
}

func (s *rancherScheduler) FrameworkMessage(
	driver sched.SchedulerDriver,
	executorID *mesos.ExecutorID,
	slaveID *mesos.SlaveID,
	message string) {
	fmt.Println("framework message = ", message)
}

func (s *rancherScheduler) OfferRescinded(_ sched.SchedulerDriver, offerID *mesos.OfferID) {
	fmt.Println("offer rescinded")
}
func (s *rancherScheduler) SlaveLost(_ sched.SchedulerDriver, slaveID *mesos.SlaveID) {
	fmt.Println("slave lost")
}
func (s *rancherScheduler) ExecutorLost(_ sched.SchedulerDriver, executorID *mesos.ExecutorID, slaveID *mesos.SlaveID, status int) {
	fmt.Println("executor lost")
}

func NewScheduler(mesosMaster string, taskChan chan *Task) chan error {
	scheduler := newRancherMesosScheduler(taskChan)
	fmt.Println(mesosMaster)
	driver, err := sched.NewMesosSchedulerDriver(sched.DriverConfig{
		Master: mesosMaster,
		Framework: &mesos.FrameworkInfo{
			Name: proto.String("RancherMesos"),
			User: proto.String("root"),
		},
		Scheduler: scheduler,
	})

	if err != nil {
		fmt.Println("ERRORRR!!!!!!")
		fmt.Println(err)
		os.Exit(0)
	}

	errChan := make(chan error, 1)
	go func() {
		fmt.Println("Running....")
		_, err := driver.Run()
		fmt.Println("skdkfnkdnfd")
		os.Exit(0)
		driver.Stop(false)
		if err != nil {
			errChan <- err
		} else {
			errChan <- nil
		}
	}()
	return errChan
}
