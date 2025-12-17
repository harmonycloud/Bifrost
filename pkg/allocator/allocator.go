// Copyright 2015 CNI authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package allocator

import (
	"fmt"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/harmonycloud/bifrost/pkg/backend"
	"github.com/harmonycloud/bifrost/pkg/backend/etcdcli"
	bflog "github.com/harmonycloud/bifrost/pkg/log"
	"github.com/harmonycloud/bifrost/pkg/types"
	"go.etcd.io/etcd/clientv3"
	"net"
)

// IPAllocator allocates IP addresses from IPPools in etcd
type IPAllocator struct {
	BackendClient backend.Backend
	etcdClient    *clientv3.Client
}

var log = bflog.Alloc

// NewIPAllocator returns a new IPAllocator from the given etcd client
func NewIPAllocator(config *types.IPAMConfig) *IPAllocator {

	if config.DatastoreType == "" {
		etcdclient, err := etcdcli.Init(config)

		if err != nil {
			log.Errorf("pool init err is %v", err)
		}
		defer func() {
			if err := etcdclient.Client.Close(); err != nil {
				log.Errorf("close etcd err %v", err)
			}
		}()

		return &IPAllocator{BackendClient: etcdclient}
	} else {
		return &IPAllocator{}
	}

}

func (alloc *IPAllocator) Close() {
	if err := alloc.etcdClient.Close(); err != nil {
		log.Errorf("close  IPAllocator err %v", err)
	}
}

// AllocateSvcSpecified returns an IP address and its svcIPPOOL if success.
func (alloc *IPAllocator) AllocateSvcSpecified(fixed bool, nameSpace, serviceIPPoolName, podName, stsName string) (*current.IPConfig, *types.NSIPPool, error) {
	var reservedIP *net.IPNet
	var reservedIPAddress *net.IP
	var err error
	var nsIPPool *types.NSIPPool

	reservedIPAddress, nsIPPool, err = alloc.BackendClient.GetServiceSpecIP(fixed, nameSpace, serviceIPPoolName, podName, stsName, alloc.etcdClient)

	if err != nil {
		return nil, nil, err
	}
	if reservedIPAddress == nil {
		return nil, nil, fmt.Errorf("no IP addresses available in serviceIPPool  %s", serviceIPPoolName)
	}
	reservedIP = &net.IPNet{IP: *reservedIPAddress, Mask: nsIPPool.Subnet.Mask}
	version := "4"
	log.Debugf("get ip succeed %v", reservedIP)
	return &current.IPConfig{
		Version: version,
		Address: *reservedIP,
	}, nsIPPool, nil
}

// Release remove an IP usage record from etcd if not fixed,
// for fixed IP, Release tags the IP with "delete" flag
func (alloc *IPAllocator) Release(nameSpace, podName string, releaseFixed bool) error {
	return alloc.BackendClient.Release(nameSpace, podName, releaseFixed)
}

func (alloc *IPAllocator) AllocateIP(nameSpace, nsIpPoolNames, podName string) (*current.IPConfig, *types.NSIPPool, error) {
	return alloc.BackendClient.AllocateIP(nameSpace, nsIpPoolNames, podName)
}
