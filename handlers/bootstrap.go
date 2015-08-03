package handlers

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/rancherio/go-machine-service/events"
	"github.com/rancherio/go-rancher/client"
	"github.com/rancherio/rancher-mesos-scheduler/tasks"
)

const (
	maxWait = time.Duration(time.Second * 10)
)

func MesosScheduleBootstrap(event *events.Event, apiClient *client.RancherClient) (err error) {
	log.WithFields(log.Fields{
		"resourceId": event.ResourceId,
		"eventId":    event.Id,
	}).Info("Bootstrapping Machine")

	machine, err := apiClient.Machine.ById(event.ResourceId)
	if err != nil {
		return handleByIdError(err, event, apiClient)
	}

	regUrl, imageRepo, imageTag, err := getRegistrationUrlAndImage(machine.AccountId, apiClient)
	if err != nil {
		return handleByIdError(err, event, apiClient)
	}

	tasks.UpdateTask(event.ResourceId, imageRepo, imageTag, regUrl, machine.ExternalId, machine.Name)

	reply := newReply(event)
	return publishReply(reply, apiClient)
}

var getRegistrationUrlAndImage = func(accountId string, apiClient *client.RancherClient) (string, string, string, error) {
	listOpts := client.NewListOpts()
	listOpts.Filters["accountId"] = accountId
	listOpts.Filters["state"] = "active"
	tokenCollection, err := apiClient.RegistrationToken.List(listOpts)
	if err != nil {
		return "", "", "", err
	}

	var token client.RegistrationToken
	if len(tokenCollection.Data) >= 1 {
		log.WithFields(log.Fields{
			"accountId": accountId,
		}).Debug("Found token for account")
		token = tokenCollection.Data[0]
	} else {
		log.WithFields(log.Fields{
			"accountId": accountId,
		}).Debug("Creating new token for account")
		createToken := &client.RegistrationToken{
			AccountId: accountId,
		}

		createToken, err = apiClient.RegistrationToken.Create(createToken)
		if err != nil {
			return "", "", "", err
		}
		createToken, err = waitForTokenToActivate(createToken, apiClient)
		if err != nil {
			return "", "", "", err
		}
		token = *createToken
	}

	regUrl, ok := token.Links["registrationUrl"]
	if !ok {
		return "", "", "", fmt.Errorf("No registration url on token [%v] for account [%v].", token.Id, accountId)
	}

	imageParts := strings.Split(token.Image, ":")
	if len(imageParts) != 2 {
		return "", "", "", fmt.Errorf("Invalid Image format in token [%v] for account [%v]", token.Id, accountId)
	}

	regUrl = tweakRegistrationUrl(regUrl)
	return regUrl, imageParts[0], imageParts[1], nil
}

func tweakRegistrationUrl(regUrl string) string {
	// We do this to accomodate end-to-end workflow in our local development environments.
	// Containers running in a vm won't be able to reach an api running on "localhost"
	// because typically that localhost is referring to the real computer, not the vm.
	localHostReplace := os.Getenv("CATTLE_AGENT_LOCALHOST_REPLACE")
	if localHostReplace == "" {
		return regUrl
	}

	regUrl = strings.Replace(regUrl, "localhost", localHostReplace, 1)
	return regUrl
}

func waitForTokenToActivate(token *client.RegistrationToken,
	apiClient *client.RancherClient) (*client.RegistrationToken, error) {
	timeoutAt := time.Now().Add(maxWait)
	ticker := time.NewTicker(time.Millisecond * 250)
	defer ticker.Stop()
	for t := range ticker.C {
		token, err := apiClient.RegistrationToken.ById(token.Id)
		if err != nil {
			return nil, err
		}
		if token.State == "active" {
			return token, nil
		}
		if t.After(timeoutAt) {
			return nil, fmt.Errorf("Timed out waiting for token to activate.")
		}
	}
	return nil, fmt.Errorf("Couldn't get active token.")
}
