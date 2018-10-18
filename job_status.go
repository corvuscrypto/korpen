package korpen

import (
	"k8s.io/api/batch/v1"
)

type Status int

const (
	Running Status = iota + 1
	Succeeded
	Failed
)

type StatusChangeHandler func(job *v1.Job)

func getStatus(job *v1.Job) Status {
	kubeStatus := job.Status
	if kubeStatus.Failed != 0 {
		return Failed
	}
	if kubeStatus.Succeeded != 0 {
		return Succeeded
	}
	return Running
}

type StatusMapper struct {
	eventChannel chan *JobEvent
	statusMap    map[string]Status
	changeFuncs  map[Status][]StatusChangeHandler
}

func NewStatusMapper(eventChan chan *JobEvent) (mapper *StatusMapper) {
	mapper = new(StatusMapper)
	mapper.statusMap = make(map[string]Status)
	mapper.eventChannel = eventChan
	mapper.changeFuncs = make(map[Status][]StatusChangeHandler)
	mapper.changeFuncs[Running] = []StatusChangeHandler{}
	mapper.changeFuncs[Succeeded] = []StatusChangeHandler{}
	mapper.changeFuncs[Failed] = []StatusChangeHandler{}
	go mapper.waitForEvents()
	return
}

func (m *StatusMapper) waitForEvents() {
	for {
		evt := <-m.eventChannel
		m.UpdateStatus(evt.Job)
	}
}

func (m *StatusMapper) UpdateStatus(job *v1.Job) {
	newStatus := getStatus(job)
	if m.statusMap[job.Name] != newStatus {
		for _, f := range m.changeFuncs[newStatus] {
			f(job)
		}
		m.statusMap[job.Name] = newStatus
	}
}

func (m *StatusMapper) AddCallback(status Status, f StatusChangeHandler) {
	m.changeFuncs[status] = append(m.changeFuncs[status], f)
}
