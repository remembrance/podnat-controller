package controller

import (
	"github.com/gutmensch/podnat-controller/internal/api"
	"github.com/gutmensch/podnat-controller/internal/common"
	"k8s.io/client-go/rest"
	"net"
	"os"
	"time"

	"golang.org/x/exp/slices"
	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/util/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type PodInformer struct {
	factory kubeinformers.SharedInformerFactory
}

func (i *PodInformer) Run() {
	stop := make(chan struct{})
	defer close(stop)
	defer runtime.HandleCrash()
	i.factory.Start(stop)
	for {
		time.Sleep(time.Second)
	}
}

func generatePodInfo(event string, data interface{}) *api.PodInfo {
	pod := data.(*corev1.Pod)
	podName := pod.ObjectMeta.Name
	podNamespace := pod.ObjectMeta.Namespace
	podAnnotation, err := api.ParseAnnotation(pod.ObjectMeta.Annotations[common.AnnotationKey])
	if err != nil {
		klog.Warningf("ignoring pod %s with invalid annotation, error: '%v'\n", podName, err)
		return nil
	}

	info := &api.PodInfo{
		Event:      event,
		Name:       podName,
		Namespace:  podNamespace,
		Node:       common.ShortHostName(pod.Spec.NodeName),
		Annotation: podAnnotation,
		IPv4:       common.ParseIP(pod.Status.PodIP),
	}
	return info
}

func filterForAnnotationAndPlacement(event string, data interface{}) bool {
	pod := data.(*corev1.Pod)

	// IP not yet assigned, wait for next update cycle
	if net.ParseIP(pod.Status.PodIP) == nil {
		return false
	}

	// not running on this node
	if common.ShortHostName(pod.Spec.NodeName) != common.NodeID {
		return false
	}

	// pod not ready (also avoids noise during pod replacement updates)
	for _, cond := range pod.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "False" && event == "update" {
			return false
		}
	}

	// valid pod and state
	if _, ok := pod.ObjectMeta.Annotations[common.AnnotationKey]; ok {
		return true
	}

	return false
}

func NewPodInformer(subscriber []string, events chan<- *api.PodInfo) *PodInformer {
	kubeConfig := common.GetEnv("KUBECONFIG", "")
	var config *rest.Config
	var clientSet *kubernetes.Clientset
	var err error
	config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		klog.Errorln(err)
		os.Exit(1)
	}
	clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorln(err)
		os.Exit(1)
	}

	in := &PodInformer{
		factory: kubeinformers.NewSharedInformerFactory(clientSet, time.Duration(common.InformerResync)*time.Second),
	}
	_, _ = in.factory.Core().V1().Pods().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if slices.Contains(subscriber, "add") && filterForAnnotationAndPlacement("add", obj) {
				pod := generatePodInfo("add", obj)
				if pod != nil {
					klog.V(9).Infof("new pod added, matched filters: %s \n", pod.Name)
					events <- pod
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			if slices.Contains(subscriber, "delete") && filterForAnnotationAndPlacement("delete", obj) {
				pod := generatePodInfo("delete", obj)
				if pod != nil {
					klog.V(9).Infof("pod deleted, matched filters: %s \n", pod.Name)
					events <- pod
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if slices.Contains(subscriber, "update") && filterForAnnotationAndPlacement("update", newObj) {
				pod := generatePodInfo("update", newObj)
				if pod != nil {
					klog.V(9).Infof("pod updated, matched filters: %s \n", pod.Name)
					events <- pod
				}
			}
		},
	})

	return in
}
