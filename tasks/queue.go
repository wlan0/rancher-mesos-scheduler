package tasks

import (
	"sync"
)

var taskList []string
var taskQueue map[string]*Task
var mutex sync.Mutex

func init() {
	taskList = []string{}
	taskQueue = map[string]*Task{}
}

func AddTask(task *Task) {
	mutex.Lock()
	defer mutex.Unlock()
	taskList = append(taskList, task.Name)
	taskQueue[task.Name] = task
}

func UpdateTask(name, imageRepo, imageTag, registrationUrl, hostUuid string) {
	mutex.Lock()
	defer mutex.Unlock()
	task, ok := taskQueue[name]
	if !ok {
		return
	}
	task.ImageRepo = imageRepo
	task.ImageTag = imageTag
	task.RegistrationUrl = registrationUrl
	task.HostUuid = hostUuid
}

func GetNextTask() *Task {
	mutex.Lock()
	defer mutex.Unlock()
	if len(taskList) == 0 {
		return nil
	}
	toReturn := taskList[0]
	if len(taskList) > 1 {
		taskList = taskList[1:]
	} else {
		taskList = []string{}
	}
	toReturnVal := taskQueue[toReturn]
	delete(taskQueue, toReturn)
	return toReturnVal
}
