package controller

import (
	"github.com/harmonycloud/bifrost/pkg/allocator"
	"github.com/harmonycloud/bifrost/pkg/types"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/klog"
)

// Retriever is used to retriever IP addresses from deleted pods
type Retriever struct {
	allocator *allocator.IPAllocator
	Clientset *kubernetes.Clientset
}

// RetrieveIP scans all pods and retrieve IP
func (r *Retriever) RetrieveIP() {
	klog.Infof("retrieve IP")
	r.retrieveNormalIP()

}

func (r *Retriever) retrieveNormalIP() {
	list, _ := r.Clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	nodes, _ := r.Clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	klog.Infof("len(nodes.Items) is %v", len(nodes.Items))
	//for _,v:=range nodes.Items{
	//	klog.Infof("node name %s",v.Name)
	//}

	//klog.Infof("list.Size() is %v", list.Size())
	klog.Infof("len(list.Items) is %v", len(list.Items))

	//for _,v:=range list.Items{
	//	klog.Infof("pod name %s",v.GetName())
	//}
	//
	//return
	klog.Infof("retrieve normal IP")
	podNetworkList, err := r.getPodList()

	if err != nil {
		klog.Errorf("get normal IP Error %s", err)
		return
	}
	ipAllocator := r.allocator

	for _, podNetwork := range podNetworkList {

		pod, err := r.Clientset.CoreV1().Pods(podNetwork.Namespace).Get(podNetwork.PodName, metav1.GetOptions{})

		if err != nil {
			if errors.IsNotFound(err) {
				klog.Infof("retriever not found pod(%s) ip", podNetwork.PodName)
				releaseError := ipAllocator.Release(podNetwork.Namespace, podNetwork.PodName, false)
				if releaseError != nil {
					klog.Errorf("retrieve normal ip for pod(%s) failed", podNetwork.PodName)
					continue
				}

			}
		} else {
			if pod.Status.PodIP != podNetwork.PodIP.String() {
				klog.Infof("podname:%s, podIp:%s,networkIp:%s ", podNetwork.PodName, pod.Status.PodIP, podNetwork.PodIP)
				releaseError := ipAllocator.Release(podNetwork.Namespace, podNetwork.PodName, false)
				if releaseError != nil {
					klog.Errorf("retrieve normal ip for pod(%s) failed", podNetwork.PodName)
					continue
				}
			}

		}

	}

}

// RetrieveFixedIP retrieves fixed IP from deleted statefulset
func (r *Retriever) RetrieveFixedIP() {
	klog.Infof("retrieve fixed IP not implement")

}

func (r *Retriever) getPodList() ([]types.PodNetwork, error) {
	return r.allocator.BackendClient.GetPodList()
}
