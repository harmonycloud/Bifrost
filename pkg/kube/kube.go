package kube

import (
	"flag"
	"fmt"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/harmonycloud/bifrost/pkg/log"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"path/filepath"
)

// Client is used to connect with k8s api server
type Client struct {
	restClient *restclient.Config
	Clientset  *kubernetes.Clientset
}

func newKubeConfig() *restclient.Config {
	kubeconfig := flag.String("kubeconfig", filepath.Join("/etc", "cni", "net.d", "hcmacvlan-kubeconfig"),
		"(optional) absolute path to the kubeconfig file")

	log.Kube.Debugf("kubeconfig is %v", *kubeconfig)
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(fmt.Errorf("err is %v, kubeconfig is %v", err, *kubeconfig))
	}

	return config
}

// NewClient returns a new k8s client if success
func NewClient() (*Client, error) {
	client := &Client{restClient: newKubeConfig()}
	if err := client.createClient(); err != nil {
		return nil, err
	}

	return client, nil
}

func (kc *Client) createClient() error {
	if kc.Clientset == nil {
		clientset, err := kubernetes.NewForConfig(kc.restClient)
		if err != nil {
			return err
		}
		kc.Clientset = clientset
	}

	return nil
}

// K8sArgs is the valid CNI_ARGS used for Kubernetes
type K8sArgs struct {
	types.CommonArgs
	K8sPodName             types.UnmarshallableString
	K8sPodNamespace        types.UnmarshallableString
	K8sPodInfraContainerID types.UnmarshallableString
	K8sAnnotation          types.UnmarshallableString
}
