package controller

import (
	"github.com/robfig/cron"
	"go.etcd.io/etcd/clientv3"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

// Cronjob uses input retriever to start an IP retrieve job,
// every period the job will be executed once.
func Cronjob(period string, r *Retriever) {
	cronjob := cron.New()
	cronjob.AddFunc(period, r.RetrieveIP)
	klog.Infof("start cron job to retrieve ip period %s", period)
	cronjob.Start()
}

// NewRetriever uses etcd and k8s client to create and return an IP retriever
func NewRetriever(clientv3 *clientv3.Client, clientset *kubernetes.Clientset) *Retriever {
	retriever := &Retriever{
		cli:       clientv3,
		Clientset: clientset,
	}
	return retriever
}
