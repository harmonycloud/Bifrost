package etcdcli

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/plugins/pkg/ip"
	bflog "github.com/harmonycloud/bifrost/pkg/log"
	"github.com/harmonycloud/bifrost/pkg/types"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	"net"
	"strings"
	"time"
)

var log = bflog.Etcd

func (etcdClient *EtcdClient) Release(nameSpace, podName string, releaseFixed bool) error {
	e := &etcdClient.Client

	response, err := concurrency.NewSTM(e, func(s concurrency.STM) error {
		podNetworkKey := fmt.Sprintf(types.Pod_Key, nameSpace, podName)
		podNetworkStr := s.Get(podNetworkKey)
		if podNetworkStr == "" {
			log.Errorf("pod  %s,%s is not exist !!!", nameSpace, podName)
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
			log.Infof("pod %s reserved fixed ip %s", podName, podNetwork.PodIP)
			return nil
		}

		var nsPoolName string
		if podNetwork.SvcIPPoolName != "" {
			servicePoolKey := fmt.Sprintf(types.Servicce_Pool_Key, nameSpace, podNetwork.SvcIPPoolName)
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
			hcPoolKey := fmt.Sprintf(types.NS_POOL_Key, nameSpace, nsPoolName)
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
			s.Put(fmt.Sprintf(types.NS_POOL_Key, nameSpace, hcPool.Name), string(hcPoolStrBytes))
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

func (alloc *EtcdClient) GetServiceSpecIP(fixed bool, nameSpace, servicePoolName, podName, stsName string, e *clientv3.Client) (*net.IP, *types.NSIPPool, error) {

	var hcPool *types.NSIPPool
	var result *net.IP
	response, err := concurrency.NewSTM(e, func(s concurrency.STM) error {
		podIPKey := fmt.Sprintf(types.Pod_Key, nameSpace, podName)
		podInfoStr := s.Get(podIPKey)
		if !fixed && podInfoStr != "" {
			return fmt.Errorf("pod %s already assain ip %v", podName, podInfoStr)

		}
		servicePoolKey := fmt.Sprintf(types.Servicce_Pool_Key, nameSpace, servicePoolName)
		str := s.Get(servicePoolKey)
		serviceIPPool := &types.ServiceIPPool{}
		err := json.Unmarshal([]byte(str), serviceIPPool)
		if err != nil {
			return fmt.Errorf("unmarshal %s to serviceIPPool err %v", str, err)
		}

		// reuse fixedIP
		if r, fixedIP, newPodInfoStr := alloc.reuseFixedIP(podInfoStr, serviceIPPool); r {
			s.Put(podIPKey, newPodInfoStr)
			log.Infof("pod %s reserved fixed ip %s", podName, fixedIP)
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
		nsPoolKey := fmt.Sprintf(types.NS_POOL_Key, nameSpace, serviceIPPool.NSIPPoolName)
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
		log.Infof("spec ip is %s,last ip is %s,pod name is %s ", ipAddr, ipAddr, podName)
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

func (alloc *EtcdClient) reuseFixedIP(podInfoStr string, pool *types.ServiceIPPool) (bool, *net.IP, string) {

	var podIP net.IP
	if podInfoStr == "" {
		log.Infof("Previous fixed ip is invalid. podinfo is %v,svc pool is %v", podInfoStr, pool)
		return false, nil, ""
	}
	var podInfo = &types.PodNetwork{}
	err := json.Unmarshal([]byte(podInfoStr), podInfo)
	if err != nil {
		log.Error(err)
		return false, nil, ""
	}

	podIP = podInfo.PodIP
	if podIP != nil && ip.Cmp(podIP, pool.Start) >= 0 && ip.Cmp(podIP, pool.End) <= 0 && podInfo.FixedIP {
		podInfo.StsDeleted = false
		podInfo.ModifyTime = time.Now()
		newPodInfoStr, err := json.Marshal(podInfo)
		if err != nil {
			log.Errorf("marshal %v to podNetworkStr %s err  %v", podInfo, newPodInfoStr, err)
			return false, nil, ""
		}
		log.Infof("Fixed ip is %s, pod name is %s ", podIP, podInfo.PodName)
		return true, &podIP, string(newPodInfoStr)
	}

	log.Infof("Previous fixed ip is invalid. podinfo is %v,svc pool is %v", podInfoStr, pool)
	return false, nil, ""

}

func (alloc *EtcdClient) AllocateIP(nameSpace, nsIpPoolNames, podName string) (*current.IPConfig, *types.NSIPPool, error) {

	for _, poolName := range strings.Split(nsIpPoolNames, ",") {
		ifconfig, nspool, err := alloc.getIPFromNSPool(nameSpace, poolName, podName)
		if err == nil && ifconfig != nil {
			return ifconfig, nspool, err
		}
	}
	return nil, nil, fmt.Errorf("nspool no ip left")

}

func (alloc *EtcdClient) getIPFromNSPool(nameSpace, poolName, podName string) (*current.IPConfig, *types.NSIPPool, error) {
	var nsIPPool *types.NSIPPool
	var result *net.IP = nil
	podIPKey := fmt.Sprintf(types.Pod_Key, nameSpace, podName)

	exchange := func(stm concurrency.STM) error {
		podInfoStr := stm.Get(podIPKey)

		if podInfoStr != "" {
			return fmt.Errorf("pod %s already assain ip %s", podName, podInfoStr)
		}
		nsIpPoolKey := fmt.Sprintf(types.NS_POOL_Key, nameSpace, poolName)
		nsIpPoolStr := stm.Get(nsIpPoolKey)
		//nsIPPool := &types.NSIPPool{}
		err := json.Unmarshal([]byte(nsIpPoolStr), &nsIPPool)
		if err != nil {
			return fmt.Errorf("unmarshal %s to nsIPPool err %v", nsIpPoolStr, err)
		}
		var cursorIp = nsIPPool.Start

		for ip.Cmp(cursorIp, nsIPPool.End) <= 0 {
			if nsIPPool.PodMap[cursorIp.String()].PodIP != nil && !nsIPPool.PodMap[cursorIp.String()].FixedIP {
				//clear leak pod
				tempPodNetworkKey := fmt.Sprintf(types.Pod_Key, nameSpace, nsIPPool.PodMap[cursorIp.String()].PodName)
				tempPpodNetworkStr := stm.Get(tempPodNetworkKey)
				//clear pod
				if tempPpodNetworkStr == "" {
					log.Infof("clear pod podip %s,podname  %s", cursorIp.String(), nsIPPool.PodMap[cursorIp.String()].PodName)
					delete(nsIPPool.PodMap, cursorIp.String())
				}
			} else {
				podNetwork := nsIPPool.PodMap[cursorIp.String()]
				podNetwork.Namespace = nameSpace
				podNetwork.PodName = podName
				podNetwork.PodIP = cursorIp
				podNetwork.NsPoolName = poolName
				podNetwork.ModifyTime = time.Now()
				podNetworkStr, err := json.Marshal(podNetwork)
				if err != nil {
					return fmt.Errorf("podnetwork  marshal err %v", err)
				}
				nsIPPool.PodMap[cursorIp.String()] = podNetwork
				nsIPPool.Used++
				hcPoolStr, err := json.Marshal(nsIPPool)
				stm.Put(nsIpPoolKey, string(hcPoolStr))
				stm.Put(podIPKey, string(podNetworkStr))
				result = &cursorIp
				return nil
			}
			cursorIp = ip.NextIP(cursorIp)
		}

		return fmt.Errorf("nspool no ip left")
	}

	response, err := concurrency.NewSTMRepeatable(context.TODO(), &alloc.Client, exchange)

	if err != nil {
		return nil, nil, err
	}

	if response.Succeeded {
		reservedIP := &net.IPNet{IP: *result, Mask: nsIPPool.Subnet.Mask}
		ipconfig := &current.IPConfig{
			Version: "4",
			Address: *reservedIP,
		}
		return ipconfig, nsIPPool, nil
	} else {
		return nil, nil, fmt.Errorf("unknown err")
	}
}
