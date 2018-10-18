package korpen

import (
	"time"

	api "k8s.io/api/batch/v1"
	"k8s.io/client-go/informers/batch/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

//go:generate stringer -type=EventType
type EventType int

const (
	JobAdded EventType = iota
	JobDeleted
	JobUpdated
)

type JobEvent struct {
	Job       *api.Job
	EventType EventType
}

type FilteredEventNotifier struct {
	watchedJobs map[string]bool
	EventChan   chan *JobEvent
}

func NewFilteredEventNotifier() *FilteredEventNotifier {
	notifier := new(FilteredEventNotifier)
	notifier.watchedJobs = make(map[string]bool)
	notifier.EventChan = make(chan *JobEvent, 10)
	return notifier
}

func (notifier *FilteredEventNotifier) handleEvent(job interface{}, evt EventType) {
	notifier.EventChan <- &JobEvent{job.(*api.Job), evt}
}

func (notifier *FilteredEventNotifier) OnAdd(obj interface{}) {
	notifier.handleEvent(obj, JobAdded)
}
func (notifier *FilteredEventNotifier) OnDelete(obj interface{}) {
	notifier.handleEvent(obj, JobDeleted)
}

func (notifier *FilteredEventNotifier) OnUpdate(oldJob, newJob interface{}) {
	notifier.handleEvent(newJob, JobUpdated)
}

func (notifier *FilteredEventNotifier) WatchJob(jobName string) {
	notifier.watchedJobs[jobName] = true
}

type Watcher struct {
	informer cache.SharedIndexInformer
	indexMap cache.Indexers
	stopChan chan struct{}
}

func NewJobWatcher(client *kubernetes.Clientset, notifier *FilteredEventNotifier) (watcher *Watcher) {
	watcher = new(Watcher)
	watcher.informer = v1.NewJobInformer(client, "default", time.Second*5, make(cache.Indexers))
	watcher.informer.AddEventHandler(notifier)
	watcher.stopChan = make(chan struct{})
	return
}

func (w *Watcher) Start() {
	go w.informer.Run(w.stopChan)
}

func (w *Watcher) Stop() {
	w.stopChan <- struct{}{}
}
