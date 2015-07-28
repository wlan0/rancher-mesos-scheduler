package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/rancherio/rancher-mesos-scheduler/events"
	"github.com/rancherio/rancher-mesos-scheduler/handlers"
)

var (
	GITCOMMIT = "HEAD"
)

func main() {
	log.WithFields(log.Fields{
		"gitcommit": GITCOMMIT,
	}).Info("Starting rancher-mesos-scheduler...")
	eventHandlers := map[string]events.EventHandler{
		"physicalhost.create": handlers.MesosSchedule,
		"ping":                handlers.PingNoOp,
	}

	apiUrl := os.Getenv("CATTLE_URL")
	accessKey := os.Getenv("CATTLE_ACCESS_KEY")
	secretKey := os.Getenv("CATTLE_SECRET_KEY")

	router, err := events.NewEventRouter("rancherMesosScheduler", 2000, apiUrl, accessKey, secretKey,
		nil, "192.168.11.210:5050", eventHandlers, 10)
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
