package korpen

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	// for GCP creds
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var SharedClient *Client

type Client struct {
	kubeClient   *kubernetes.Clientset
	watchedTasks []*Task
	started      bool
	statusCache  map[string]TaskStatus
}

func (k *Client) watchJobs() {
	k.statusCache = make(map[string]TaskStatus)
	for {
		for _, task := range k.watchedTasks {
			cachedStatus, found := k.statusCache[task.PodID]
			status := k.GetStatus(task)
			if !found || cachedStatus != status {
				fmt.Println("Status Changed on Pod", task.PodID, "to status", status)
				k.statusCache[task.PodID] = status
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func (k *Client) WatchTask(t *Task) {
	k.watchedTasks = append(k.watchedTasks, t)
	if !k.started {
		k.started = true
		go k.watchJobs()
	}
}

func (k *Client) GetStatus(t *Task) TaskStatus {
	job, err := k.kubeClient.BatchV1().Jobs("default").Get(t.PodID, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	if job.Status.Active != 0 {
		return TaskRunning
	} else if job.Status.Failed != 0 {
		return TaskFailed
	} else if job.Status.Succeeded != 0 {
		return TaskSucceeded
	}
	return "unknown"
}

func GetKubeConfig() *rest.Config {
	usr, _ := user.Current()
	cfgPath, ok := os.LookupEnv("KUBERNETES")
	if !ok {
		cfgPath = path.Join(usr.HomeDir, ".kube/config")
	}
	config, err := clientcmd.BuildConfigFromFlags("", cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

// InitClient inits the client for shared use
func InitClient() {
	config := GetKubeConfig()
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	SharedClient = &Client{}
	SharedClient.kubeClient = clientSet
}
