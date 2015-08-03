package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/rancherio/go-machine-service/events"
	"github.com/rancherio/rancher-mesos-scheduler/handlers"
	"github.com/rancherio/rancher-mesos-scheduler/scheduler"
)

var (
	GITCOMMIT = "HEAD"
)

func main() {
	log.WithFields(log.Fields{
		"gitcommit": GITCOMMIT,
	}).Info("Starting rancher-mesos-scheduler...")

	eventHandlers := map[string]events.EventHandler{
		"physicalhost.create":    handlers.MesosScheduleCreate,
		"physicalhost.bootstrap": handlers.MesosScheduleBootstrap,
		"ping": handlers.PingNoOp,
	}

	apiUrl := os.Getenv("CATTLE_URL")
	accessKey := os.Getenv("CATTLE_ACCESS_KEY")
	secretKey := os.Getenv("CATTLE_SECRET_KEY")
	mesosMaster := os.Getenv("MESOS_MASTER")
	bridgeCIDR := os.Getenv("IP_CIDR")

	go func() {
		err := scheduler.NewScheduler(mesosMaster, bridgeCIDR)
		log.WithFields(log.Fields{
			"err": err,
		}).Info("Finished Processing. Exiting.")
		os.Exit(0)
	}()

	router, err := events.NewEventRouter("rancherMesosScheduler", 2000, apiUrl, accessKey, secretKey,
		nil, eventHandlers, 10)
	if err != nil {
		log.WithFields(log.Fields{
			"Err": err,
		}).Error("Unable to create EventRouter")
	} else {
		err := router.Start(nil)
		if err != nil {
			log.WithFields(log.Fields{
				"Err": err,
			}).Error("Unable to start EventRouter")
		}
	}
	log.Info("Exiting rancher-mesos-scheduler...")
}
