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
	"context"
	"encoding/json"
	"fmt"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/harmonycloud/hcbridge/pkg/log"
	"github.com/harmonycloud/hcbridge/pkg/types"
	"go.etcd.io/etcd/clientv3"
	"net"
	"time"
)

// IPAllocator allocates IP addresses from IPPools in etcd
type IPAllocator struct {
	etcdClient *clientv3.Client
}

// NewIPAllocator returns a new IPAllocator from the given etcd client
func NewIPAllocator(etcdClient *clientv3.Client) *IPAllocator {
	return &IPAllocator{
		etcdClient: etcdClient,
	}
}

// AllocateIP returns an IP address and its nsIPPool if success.
func (alloc *IPAllocator) AllocateIP(fixed bool, nameSpace, serviceIPPoolName, podName, stsName string) (*current.IPConfig, *types.NSIPPool, error) {
	var reservedIP *net.IPNet
	var reservedIPAddress *net.IP
	var err error
	var nsIPPool *types.NSIPPool

	reservedIPAddress, nsIPPool, err = alloc.getServiceSpecIP(fixed, nameSpace, serviceIPPoolName, podName, stsName, alloc.etcdClient)

	if err != nil {
		return nil, nil, err
	}
	if reservedIPAddress == nil {
		return nil, nil, fmt.Errorf("no IP addresses available in serviceIPPool  %s", serviceIPPoolName)
	}
	reservedIP = &net.IPNet{IP: *reservedIPAddress, Mask: nsIPPool.Subnet.Mask}
	version := "4"
	log.Alloc.Debugf("get ip succeed %v", reservedIP)
	return &current.IPConfig{
		Version: version,
		Address: *reservedIP,
	}, nsIPPool, nil
}

// Release remove an IP usage record from etcd if not fixed,
// for fixed IP, Release tags the IP with "delete" flag
func (alloc *IPAllocator) Release(nameSpace, podName string, releaseFixed bool) error {
	e := alloc.etcdClient

	response, err := concurrency.NewSTMRepeatable(context.TODO(), e, func(s concurrency.STM) error {
		podNetworkKey := fmt.Sprintf(types.PodKey, nameSpace, podName)
		podNetworkStr := s.Get(podNetworkKey)
		if podNetworkStr == "" {
			log.Alloc.Errorf("pod  %s,%s is not exist !!!", nameSpace, podName)
			return nil
		}
		podNetwork := &types.PodNetwork{}
		err := json.Unmarshal([]byte(podNetworkStr), podNetwork)
		if err != nil {
			return fmt.Errorf("unmarshal %s to podNetwork %v err  %v", podNetworkStr, podNetwork, err)
		}

		// reserve fixed ip
		if podNetwork.FixedIP && !releaseFixed {
			podNetwork.StsDeleted = true
			podNetwork.ModifyTime = time.Now()
			podNetworkStr, err := json.Marshal(podNetwork)
			if err != nil {
				return fmt.Errorf("marshal %v to podNetworkStr %s err  %v", podNetwork, podNetworkStr, err)
			}
			s.Put(podNetworkKey, string(podNetworkStr))
			log.Alloc.Infof("pod %s reserved fixed ip %s", podName, podNetwork.PodIP)
			return nil
		}

		var nsPoolName string
		if podNetwork.SvcIPPoolName != "" {
			servicePoolKey := fmt.Sprintf(types.ServiccePoolKey, nameSpace, podNetwork.SvcIPPoolName)
			str := s.Get(servicePoolKey)
			serviceIPPool := &types.ServiceIPPool{}
			err := json.Unmarshal([]byte(str), serviceIPPool)
			if err != nil {
				return fmt.Errorf("unmarshal %s to serviceIPPool err ", str)
			}
			serviceIPPool.Used--

			svcPoolStr, err := json.Marshal(serviceIPPool)
			if err != nil {
				return fmt.Errorf("svcPoolStr marshal err %v", err)
			}
			s.Put(servicePoolKey, string(svcPoolStr))
			nsPoolName = serviceIPPool.NSIPPoolName

		} else {
			return nil
		}
		if nsPoolName != "" {
			hcPoolKey := fmt.Sprintf(types.NSPoolKey, nameSpace, nsPoolName)
			hcPoolStr := s.Get(hcPoolKey)
			hcPool := &types.NSIPPool{}
			err := json.Unmarshal([]byte(hcPoolStr), hcPool)

			if err != nil {
				return fmt.Errorf("unmarshal %s to hcPool err %v ", hcPoolStr, err)
			}

			//delete(hcPool.PodMap, podNetwork.PodIP.String())

			temPod := types.PodNetwork{}
			oldPod := hcPool.PodMap[podNetwork.PodIP.String()]
			//temPod.SvcName = ""
			temPod.PodIP = oldPod.PodIP
			temPod.SvcName = oldPod.SvcName
			temPod.SvcIPPoolName = oldPod.SvcIPPoolName
			temPod.Namespace = oldPod.Namespace
			temPod.StsDeleted = false

			hcPool.PodMap[podNetwork.PodIP.String()] = temPod
			hcPoolStrBytes, err := json.Marshal(hcPool)

			if err != nil {
				return fmt.Errorf("hcPoolS  marshal err %v", err)
			}
			s.Put(fmt.Sprintf(types.NSPoolKey, nameSpace, hcPool.Name), string(hcPoolStrBytes))
		}
		s.Del(podNetworkKey)
		return nil
	})
	if err != nil {
		return err
	}
	if response.Succeeded {
		return nil
	}
	return fmt.Errorf("unknown err")
}

func (alloc *IPAllocator) getServiceSpecIP(fixed bool, nameSpace, servicePoolName, podName, stsName string, e *clientv3.Client) (*net.IP, *types.NSIPPool, error) {

	var hcPool *types.NSIPPool
	var result *net.IP
	response, err := concurrency.NewSTMRepeatable(context.TODO(), e, func(s concurrency.STM) error {
		podIPKey := fmt.Sprintf(types.PodKey, nameSpace, podName)
		podInfoStr := s.Get(podIPKey)
		if !fixed && podInfoStr != "" {
			return fmt.Errorf("pod %s already assain ip %v", podName, podInfoStr)

		}
		servicePoolKey := fmt.Sprintf(types.ServiccePoolKey, nameSpace, servicePoolName)
		str := s.Get(servicePoolKey)
		serviceIPPool := &types.ServiceIPPool{}
		err := json.Unmarshal([]byte(str), serviceIPPool)
		if err != nil {
			return fmt.Errorf("unmarshal %s to serviceIPPool err %v", str, err)
		}

		// reuse fixedIP
		if r, fixedIP, newPodInfoStr := alloc.reuseFixedIP(podInfoStr, serviceIPPool); r {
			s.Put(podIPKey, newPodInfoStr)
			log.Alloc.Infof("pod %s reserved fixed ip %s", podName, fixedIP)
			result = fixedIP
			return nil
		}

		//check total<serviceUsedCount

		if serviceIPPool.Used >= serviceIPPool.Total {
			return fmt.Errorf("ippoolcli %s has no ip left !!! serviceUsedCount is %d,total is %d", servicePoolName, serviceIPPool.Used, serviceIPPool.Total)
		}

		//finish check total <serviceUsedCount

		var ipAddr = serviceIPPool.LastAssignedIP
		if ipAddr == nil || ip.Cmp(ipAddr, serviceIPPool.Start) < 0 || ip.Cmp(ipAddr, serviceIPPool.End) > 0 {
			ipAddr = serviceIPPool.Start
		} else {
			ipAddr = ip.NextIP(ipAddr)
			if ip.Cmp(ipAddr, serviceIPPool.End) > 0 {
				ipAddr = serviceIPPool.Start
			}
		}
		nsPoolKey := fmt.Sprintf(types.NSPoolKey, nameSpace, serviceIPPool.NSIPPoolName)
		nsPoolStr := s.Get(nsPoolKey)
		// If the ip is used, check next ip
		hcPool = &types.NSIPPool{}
		err = json.Unmarshal([]byte(nsPoolStr), hcPool)
		if err != nil {
			return fmt.Errorf("unmarshal %s to hcPool err %v ", str, err)
		}

		var flag = false
		podNetworkMap := hcPool.PodMap
		podNetwork, _ := podNetworkMap[ipAddr.String()]
		for podNetwork.PodName != "" {
			var nextIP net.IP
			if serviceIPPool.End.Equal(ipAddr) {
				if flag {
					return fmt.Errorf("hcPool %s has no ip left !!! hcPool is %d,total is %d", hcPool.Name, hcPool.Used, hcPool.Total)
				}
				nextIP = serviceIPPool.Start
				flag = true
			} else {
				nextIP = ip.NextIP(ipAddr)
			}

			ipAddr = nextIP
			podNetwork, _ = podNetworkMap[ipAddr.String()]
		}

		/*podNetwork := types.PodNetwork{
			PodIP:         ipAddr,
			PodName:       podName,
			SvcName:       serviceIPPool.Name,
			SvcIPPoolName: servicePoolName,
			Namespace:     nameSpace,
		}*/
		podNetwork.PodName = podName
		podNetwork.SvcName = serviceIPPool.ServiceName
		if fixed {
			podNetwork.StsName = stsName
			podNetwork.StsDeleted = false
			podNetwork.FixedIP = true
			podNetwork.ModifyTime = time.Now()
		}
		hcPool.PodMap[ipAddr.String()] = podNetwork

		// Record pod:IP, podNetwork
		//podIPKey := fmt.Sprintf(types.Pod_Key, nameSpace, podName)
		podNetworkStr, err := json.Marshal(podNetwork)
		if err != nil {
			return fmt.Errorf("podnetwork  marshal err %v", err)
		}
		s.Put(podIPKey, string(podNetworkStr))

		serviceIPPool.Used++
		serviceIPPool.LastAssignedIP = ipAddr
		svcPoolStr, err := json.Marshal(serviceIPPool)

		if err != nil {
			return fmt.Errorf("svcPoolStr  marshal err %v", err)
		}

		s.Put(servicePoolKey, string(svcPoolStr))
		hcPoolStr, err := json.Marshal(hcPool)

		if err != nil {
			return fmt.Errorf("hcPoolS  marshal err %v", err)
		}
		//
		s.Put(nsPoolKey, string(hcPoolStr))

		result = &ipAddr
		log.Alloc.Infof("spec ip is %s,last ip is %s,pod name is %s ", ipAddr, ipAddr, podName)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	if response.Succeeded {
		return result, hcPool, nil
	}
	return nil, hcPool, fmt.Errorf("unknown err")
}

func (alloc *IPAllocator) reuseFixedIP(podInfoStr string, pool *types.ServiceIPPool) (bool, *net.IP, string) {
	var podInfo *types.PodNetwork
	var podIP net.IP

	if podInfoStr != "" {
		err := json.Unmarshal([]byte(podInfoStr), podInfo)
		if err != nil {
			log.Alloc.Error(err)
			return false, nil, ""
		}
		podIP = podInfo.PodIP
		if podIP != nil && ip.Cmp(podIP, pool.Start) >= 0 && ip.Cmp(podIP, pool.End) <= 0 && podInfo.FixedIP {
			podInfo.StsDeleted = false
			podInfo.ModifyTime = time.Now()
			newPodInfoStr, err := json.Marshal(podInfo)
			if err != nil {
				log.Alloc.Errorf("marshal %v to podNetworkStr %s err  %v", podInfo, newPodInfoStr, err)
				return false, nil, ""
			}
			log.Alloc.Infof("Fixed ip is %s, pod name is %s ", podIP, podInfo.PodName)
			return true, &podIP, string(newPodInfoStr)
		}
		log.Alloc.Infof("Previous fixed ip is invalid. IP is %s,"+
			" pod name is %s , pool name is %s", podIP, podInfo.PodName, pool.Name)
	}

	return false, nil, ""
}
