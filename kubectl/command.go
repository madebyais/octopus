package kubectl

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

// ICommand ...
type ICommand interface {
	GetAll(namespace string) ([]map[string]string, error)
	GetByPrefix(namespace string, prefix string) ([]map[string]string, error)
}

// Command ...
type Command struct {
	Client *kubernetes.Clientset
}

// New initiates new instance for kubectl
func New() ICommand {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return &Command{
		Client: clientset,
	}
}

// GetAll ...
func (cmd *Command) GetAll(namespace string) ([]map[string]string, error) {
	var items []map[string]string

	data, err := cmd.Client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return items, err
	}

	for _, item := range data.Items {
		tempItem := make(map[string]string)
		tempItem["namespace"] = fmt.Sprintf("\033[1;32m%s\033[0m", item.GetNamespace())
		tempItem["node"] = item.Spec.NodeName
		tempItem["serviceId"] = item.GetName()
		tempItem["serviceAddress"] = item.Status.PodIP
		items = append(items, tempItem)
	}

	return items, nil
}

// GetByPrefix ...
func (cmd *Command) GetByPrefix(namespace string, prefix string) ([]map[string]string, error) {
	var items []map[string]string

	data, err := cmd.Client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return items, err
	}

	for _, item := range data.Items {
		r, _ := regexp.Compile("^" + prefix)
		isMatch := r.MatchString(item.GetName())
		if isMatch {
			tempItem := make(map[string]string)

			namespace := "default"
			if item.GetNamespace() != namespace {
				namespace = fmt.Sprintf("\033[1;32m%s\033[0m", item.GetNamespace())
			}

			tempItem["namespace"] = namespace
			tempItem["node"] = item.Spec.NodeName
			tempItem["serviceId"] = item.GetName()
			tempItem["serviceAddress"] = item.Status.PodIP
			items = append(items, tempItem)
		}
	}

	return items, nil
}

func (cmd *Command) log(text string) {
	fmt.Printf("\n[ KUBECTL ] %s", text)
	fmt.Println(``)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
