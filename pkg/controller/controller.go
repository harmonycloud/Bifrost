package controller

import (
	"github.com/harmonycloud/bifrost/pkg/allocator"
	"github.com/robfig/cron"
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
func NewRetriever(alloc *allocator.IPAllocator, clientset *kubernetes.Clientset) *Retriever {
	retriever := &Retriever{
		allocator: alloc,
		Clientset: clientset,
	}
	return retriever
}
