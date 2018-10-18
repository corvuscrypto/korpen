package korpen

import (
	"log"
	"os"
	"os/user"
	"path"

	"k8s.io/client-go/kubernetes"
	// for GCP creds
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

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

// GetClient inits the client for shared use
func GetClient() *kubernetes.Clientset {
	config := GetKubeConfig()
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	return clientSet
}
