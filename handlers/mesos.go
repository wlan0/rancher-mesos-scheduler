package handlers

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/rancherio/go-rancher/client"
	"github.com/rancherio/rancher-mesos-scheduler/events"
	"github.com/rancherio/rancher-mesos-scheduler/mesos"
)

func MesosSchedule(event *events.Event, apiClient *client.RancherClient, taskChan chan<- *mesos.Task) (err error) {
	log.WithFields(log.Fields{
		"resourceId": event.ResourceId,
		"eventId":    event.Id,
	}).Info("Creating Machine")

	task := &mesos.Task{Id: event.Id, Name: event.ResourceId}
	taskChan <- task
	reply := newReply(event)

	end := make(chan bool, 1)
	done := make(chan bool, 1)

	go func() {
		for {
			select {
			case <-end:
				return
			case <-time.After(100 * time.Millisecond):
				if task.IsScheduled() {
					log.WithFields(log.Fields{
						"resourceId": event.ResourceId,
						"eventId":    event.Id,
					}).Info("Task has been scheduled.")
					done <- true
					return
				}
				continue
			}
		}
	}()

	select {
	case <-done:
		return publishReply(reply, apiClient)
	case <-time.After(10 * time.Second):
		end <- false
		return nil
	}

	return publishReply(reply, apiClient)
}
