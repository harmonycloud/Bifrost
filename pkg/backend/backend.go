package backend

import (
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/harmonycloud/bifrost/pkg/types"
	"net"
)

type Backend interface {
	//getIPFromNSPool(nameSpace, poolName, podName string) (*current.IPConfig, *types.NSIPPool, error)/
	AllocateIP(nameSpace, nsIpPoolNames, podName string) (*current.IPConfig, *types.NSIPPool, error)

	Release(nameSpace, podName string, releaseFixed bool) error

	GetServiceSpecIP(fixed bool, nameSpace, servicePoolName, podName, stsName string) (*net.IP, *types.NSIPPool, error)

	GetPodList() ([]types.PodNetwork, error)

	CreateIPPool(pool *types.NSIPPool) error

	GetIPPoolByName(nsname, poolName string) *types.NSIPPool

	GetNSAllPool(namespace string) []*types.NSIPPool

	GetServicePool(namespace string, ServicePoolName string) (*types.ServiceIPPool, error)

	GetAllPool() []*types.NSIPPool

	Delete(pool *types.NSIPPool) error

	CreateServiceIPPool(serviceIPPool *types.ServiceIPPool) error

	DeleteServiceIPPool(namespace string, servicePoolName string) error

	GetService(namespace, serviceName string) *types.CreateServiceRequest
}
