package kube

import (
	"fmt"
	"github.com/harmonycloud/bifrost/pkg/log"
	"github.com/harmonycloud/bifrost/pkg/types"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetK8sPodAnnotations returns the pod's annotation through k8s api server
func (kc *Client) GetK8sPodAnnotations(k8sPodNamespaces string, k8sPodName string) (map[string]string,
	error) {
	pod, err := kc.getPodDefinition(k8sPodNamespaces, k8sPodName)

	if err != nil {
		log.Kube.Errorf("can not get pod %s", k8sPodName)
		return nil, err
	}
	log.Kube.Debugf("%v", *pod)
	return pod.Annotations, nil
}

// GetPodDefinition gets pod definition through k8s api server
func (kc *Client) getPodDefinition(podNamespace string, podName string) (*v1.Pod, error) {
	client := kc.Clientset
	log.Kube.Infof("pod namespaces is %s ,pod name is %s ", podNamespace, podName)
	pod, err := client.CoreV1().Pods(podNamespace).Get(fmt.Sprintf("%s", podName), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pod, nil
}

// IsPodFixedIP returns true for pods in the control of statefulset,
// and with the annotation "hcipam_ip_fixed=true"
func (kc *Client) IsPodFixedIP(podNamespace string, podName string) bool {
	pod, err := kc.getPodDefinition(podNamespace, podName)
	if err != nil {
		log.Kube.Errorf("failed to get pod: %v", err)
		return false
	}
	value, ok := pod.Annotations[types.POD_IP_FIXED]
	if ok && pod.OwnerReferences != nil {
		reference := pod.OwnerReferences[0]
		isStatefulSet := reference.Kind == "StatefulSet"
		return value == "true" && isStatefulSet
	}
	return false
}

// GetStsName returns the name of the pod's owner
func (kc *Client) GetStsName(podNamespace string, podName string) string {
	pod, err := kc.getPodDefinition(podNamespace, podName)
	if err != nil {
		log.Kube.Errorf("failed to get pod: %v", err)
		return ""
	}

	if pod.OwnerReferences != nil {
		reference := pod.OwnerReferences[0]
		return reference.Name
	}
	log.Kube.Errorf("failed to get statefulset name")
	return ""
}
