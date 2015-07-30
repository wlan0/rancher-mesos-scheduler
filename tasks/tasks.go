package tasks

import (
	"encoding/json"
)

type Task struct {
	Id              string  `json: "-"`
	Name            string  `json: "-"`
	CPU             float64 `json: "-"`
	Mem             float64 `json: "-"`
	RegistrationUrl string  `json: "registration_url"`
	ImageRepo       string  `json: "image_repo"`
	ImageTag        string  `json: "image_tag"`
	HostUuid        string  `json: "host_uuid"`
}

func (t *Task) Marshal() ([]byte, error) {
	return json.Marshal(t)
}
