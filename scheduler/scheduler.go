package scheduler

import (
	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	mesos "github.com/mesos/mesos-go/mesosproto"
	"github.com/mesos/mesos-go/mesosutil"
	sched "github.com/mesos/mesos-go/scheduler"
	"github.com/rancherio/rancher-mesos-scheduler/tasks"
)

const (
	taskCPUs = 1
	taskMem  = 1024
	// make this configurable
	cmd = "rancher-mesos-executor --work_dir=/home/wlan0/Media/hdd"
)

var (
	defaultFilter = &mesos.Filters{
		RefuseSeconds: proto.Float64(1),
	}
)

type rancherScheduler struct {
	rancherExecutor *mesos.ExecutorInfo
}

func (s *rancherScheduler) Registered(_ sched.SchedulerDriver, frameworkID *mesos.FrameworkID, masterInfo *mesos.MasterInfo) {
	log.Info("Registered Rancher Scheduler")
}

func (s *rancherScheduler) Reregistered(_ sched.SchedulerDriver, masterInfo *mesos.MasterInfo) {
	log.Info("Re-Registered Rancher Scheduler")
}

func (s *rancherScheduler) Disconnected(sched.SchedulerDriver) {
	log.Info("Framework disconnected from master")
}

func (s *rancherScheduler) ResourceOffers(driver sched.SchedulerDriver, offers []*mesos.Offer) {
	task := tasks.GetNextTask()
	if task == nil {
		for _, of := range offers {
			driver.DeclineOffer(of.Id, defaultFilter)
		}
		return
	}
	if task.RegistrationUrl == "" {
		tasks.AddTask(task)
		for _, of := range offers {
			driver.DeclineOffer(of.Id, defaultFilter)
		}
		return
	}
	taskBytes, err := task.Marshal()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Error Marshalling task")
		for _, of := range offers {
			driver.DeclineOffer(of.Id, defaultFilter)
		}
		return
	}
	for _, offer := range offers {
		inadequate := false
		for _, res := range offer.GetResources() {
			if res.GetName() == "cpus" && *res.GetScalar().Value < 1 {
				driver.DeclineOffer(offer.Id, defaultFilter)
				inadequate = true
				continue
			}
			if res.GetName() == "mem" && *res.GetScalar().Value < 1024 {
				driver.DeclineOffer(offer.Id, defaultFilter)
				inadequate = true
				continue
			}
		}
		if inadequate {
			continue
		}
		mesosTask := &mesos.TaskInfo{
			TaskId: &mesos.TaskID{
				Value: proto.String(task.HostUuid),
			},
			SlaveId: offer.SlaveId,
			Resources: []*mesos.Resource{
				mesosutil.NewScalarResource("cpus", taskCPUs),
				mesosutil.NewScalarResource("mem", taskMem),
			},
			Data:     taskBytes,
			Name:     &task.Name,
			Executor: s.rancherExecutor,
		}
		driver.LaunchTasks([]*mesos.OfferID{offer.Id}, []*mesos.TaskInfo{mesosTask}, defaultFilter)
	}
}

func (s *rancherScheduler) Error(_ sched.SchedulerDriver, err string) {
	log.WithFields(log.Fields{
		"err": err,
	}).Info("Error while scheduling")
}

func (s *rancherScheduler) StatusUpdate(driver sched.SchedulerDriver, status *mesos.TaskStatus) {
	log.WithFields(log.Fields{
		"status": status,
	}).Info("Task status update received")
}

func (s *rancherScheduler) FrameworkMessage(driver sched.SchedulerDriver, executorID *mesos.ExecutorID, slaveID *mesos.SlaveID, message string) {
	log.WithFields(log.Fields{
		"message": message,
	}).Info("Framework message received")
}

func (s *rancherScheduler) OfferRescinded(_ sched.SchedulerDriver, offerID *mesos.OfferID) {
	log.WithFields(log.Fields{
		"offerId": offerID,
	}).Info("Offer rescinded by master")
}
func (s *rancherScheduler) SlaveLost(_ sched.SchedulerDriver, slaveID *mesos.SlaveID) {
	log.WithFields(log.Fields{
		"slaveId": slaveID,
	}).Info("Slave lost")
}
func (s *rancherScheduler) ExecutorLost(_ sched.SchedulerDriver, executorID *mesos.ExecutorID, slaveID *mesos.SlaveID, status int) {
	log.WithFields(log.Fields{
		"executorId": executorID,
		"slaveId":    slaveID,
		"status":     status,
	}).Info("Executor lost")
}

func NewScheduler(mesosMaster string) error {
	log.Info("Rancher Mesos Scheduler started")
	scheduler := &rancherScheduler{
		rancherExecutor: &mesos.ExecutorInfo{
			Command: &mesos.CommandInfo{
				Value: proto.String(cmd),
				Uris:  []*mesos.CommandInfo_URI{},
			},
			ExecutorId: &mesos.ExecutorID{Value: proto.String("rancher-mesos-executor")},
			Name:       proto.String("RancherExecutor"),
		},
	}
	driver, err := sched.NewMesosSchedulerDriver(sched.DriverConfig{
		Master: mesosMaster,
		Framework: &mesos.FrameworkInfo{
			Name: proto.String("RancherMesos"),
			User: proto.String("root"), // make this configurable
		},
		Scheduler: scheduler,
	})

	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Error starting scheduler driver")
		return err
	}

	_, err = driver.Run()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Error Running driver")
		return err
	}
	_, err = driver.Stop(false)
	return err
}
