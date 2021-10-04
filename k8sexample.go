package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PodReq struct {
	Namespace   string
	ServiceName string
	Port        string
	stopChan    chan struct{}
	readyChan   chan struct{}
}

func forwardAndWork() {
	ctx := context.Background()

	//create an in memory config and clientset
	config, err := createConfig()
	handlek8serror(err)

	//create the client set
	clientset, err := defaultClientSet(config)
	handlek8serror(err)

	//set up channels to communicate with the port forwards
	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{}, 1)

	wg := *&sync.WaitGroup{}
	wg.Add(1)
	//set up a manager goroutine for our port forward
	go func() {
		for {
			select {
			case <-readyChan:
				//port forward is ready, do your work
				doWork(stopChan)
			case <-stopChan:
				wg.Done()
				return //our defers will handle closing
			}
		}
	}()
	defer close(stopChan)
	defer close(readyChan)

	podToForward := PodReq{
		Namespace:   "data",         //"YOUR_NAMESPACE",
		ServiceName: "data-service", //"YOUR_SERVICENAME",
		Port:        "999:8080",     //"YOUR_PORT",
		stopChan:    stopChan,
		readyChan:   readyChan,
	}

	err = newPortForwarder(ctx, config, clientset, podToForward)
	handlek8serror(err)

	wg.Wait()
}

func doWork(stopChan chan struct{}) {
	//sleeping is work
	time.Sleep(10 * time.Second)

	//work is done, send signal to stop channel
	stopChan <- struct{}{}
}

//This function tries to find where your in filesystem kubeconfig is, then creates an
//in memory kubeconfig which can be used in the rest of this program execution
func createConfig() (*rest.Config, error) {
	//find home directory
	home, exists := os.LookupEnv("HOME")
	if !exists {
		home = "~"
	}

	//assume that ~/.kube/config exists and is the location of the kube config yaml file
	configPath := filepath.Join(home, ".kube", "config")

	//build an in memory config so we can connect to the cluster using the client
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

//Create an in memory clientset which is magic and lets us do all sorts of great stuff
func defaultClientSet(cfg *rest.Config) (clientcorev1.CoreV1Interface, error) {
	var config *rest.Config
	var err error

	//if the passed in config does not exist, create an in memory config
	if cfg == nil {
		config, err = createConfig()
		if err != nil {
			return nil, err
		}
	}

	//create the client we will use for all kubernetes commands
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset.CoreV1(), nil
}

//This function grabs an available pod on our servive, forwards a local port to the pod, and this sends back a struct
//on the pod.readyChan when the connection has been made. It does listen to stopChan, the closing of which will close the connection
func newPortForwarder(ctx context.Context, config *rest.Config, clientset clientcorev1.CoreV1Interface, pod PodReq) error {
	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		log.Println("error setting up round tripper")
		return err
	}

	//This is all config values and variables
	ports := []string{pod.Port}
	podName, err := getPod(ctx, clientset, pod) //TODO - finish finding pod
	if err != nil {
		return err
	}
	fmt.Printf("name of pod: %s\n", podName)
	path := fmt.Sprintf("api/v1/namespaces/%s/pods/%s/portforward", pod.Namespace, podName)
	hostIP := strings.TrimLeft(config.Host, "htps:/")
	serverUrl := url.URL{Scheme: "https", Path: path, Host: hostIP}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverUrl)

	out, errOut := new(bytes.Buffer), new(bytes.Buffer)

	//This is the actual forwarder
	forwarder, err := portforward.New(dialer, ports, pod.stopChan, pod.readyChan, out, errOut)
	if err != nil {
		//log and do the bad
		log.Fatalf("error creating port forwarder: %s", err.Error())
		return err
	}

	//forward ports
	if err = forwarder.ForwardPorts(); err != nil {
		log.Fatalf("error with the port forwarder: %s", err.Error())
		return err
	}

	return nil
}

func getPod(ctx context.Context, clientSet clientcorev1.CoreV1Interface, podCfg PodReq) (string, error) {
	fmt.Printf("incoming port-forward request: %+v\n", podCfg)
	var mySvc corev1.Service

	//get the service which has the pod
	//TODO - This can be a func
	svcListOptions := metav1.ListOptions{}
	svcs, err := clientSet.Services(podCfg.Namespace).List(ctx, svcListOptions)
	if err != nil {
		fmt.Println("error getting services list")
		return "", err
	}
	fmt.Printf("services returned: %+v\n", svcs.Items)
	for _, svc := range svcs.Items {
		if svc.Name == podCfg.ServiceName {
			log.Printf("\nwe do get a service name: %s\n", svc.Name)
			mySvc = svc
			break
		}
	}

	//get the pods for the service
	//TODO - this can be a func
	set := labels.Set(mySvc.Spec.Selector)
	podListOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	pods, err := clientSet.Pods(podCfg.Namespace).List(ctx, podListOptions)
	if err != nil {
		fmt.Println("error getting pods")
		return "", err
	}

	//TODO - clean all this up
	for _, pod := range pods.Items {
		fmt.Printf("Pod Name: %v\n", pod.Name)
		if pod.Status.Phase == corev1.PodPhase(corev1.PodRunning) {
			for _, cond := range pod.Status.Conditions {
				if (cond.Type == "ContainersReady" || cond.Type == "Ready") && cond.Status == "True" {
					return pod.GetName(), nil
				}
			}
		}
	}

	//get a ready port on the pod

	return "", nil

}

func handlek8serror(err error) {
	if err != nil {
		log.Panicf("error with kubernetes command: %s", err.Error)
	}
}
