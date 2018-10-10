package korpen

import "time"

type TaskSpec struct {
	ImageName string
	Command   string
	CPU       string
	Memory    string
}

// TaskStatus is a namespaced string
type TaskStatus string

// Statuses expected for jobs
const (
	TaskRunning   TaskStatus = "Running"
	TaskSucceeded            = "Succeeded"
	TaskFailed               = "Failed"
)

type TaskDetailsResponse struct {
	Status TaskStatus
}

type Task struct {
	PodID   string
	Spec    *TaskSpec
	Created time.Time
}
