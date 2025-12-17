package backend

import (
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/harmonycloud/bifrost/pkg/types"
	"go.etcd.io/etcd/clientv3"
	"net"
)

type Backend interface {
	//getIPFromNSPool(nameSpace, poolName, podName string) (*current.IPConfig, *types.NSIPPool, error)/
	AllocateIP(nameSpace, nsIpPoolNames, podName string) (*current.IPConfig, *types.NSIPPool, error)
	Release(nameSpace, podName string, releaseFixed bool) error
	GetServiceSpecIP(fixed bool, nameSpace, servicePoolName, podName, stsName string, e *clientv3.Client) (*net.IP, *types.NSIPPool, error)
}
