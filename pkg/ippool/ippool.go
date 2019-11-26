package ippool

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/harmonycloud/hcbridge/pkg/etcdcli"
	"github.com/harmonycloud/hcbridge/pkg/log"
	hcipamTypes "github.com/harmonycloud/hcbridge/pkg/types"
	"github.com/harmonycloud/hcbridge/pkg/utils"
	"go.etcd.io/etcd/clientv3"
)

// IPPoolClient controls all IPPools' creation and deletion in etcd
type IPPoolClient struct {
	Cli *clientv3.Client
}

// NewIPPoolClient returns a new IPPoolClient from the given config
func NewIPPoolClient(etcdClient *clientv3.Client) *IPPoolClient {
	return &IPPoolClient{Cli: etcdClient}
}

// CreateIPPool calls PutKV to create ns IPPool in etcd
func (client *IPPoolClient) CreateIPPool(pool *hcipamTypes.NSIPPool) error {
	podMAP := make(map[string]hcipamTypes.PodNetwork)
	pool.PodMap = podMAP
	pool.Total = utils.GetTotal(*pool)
	hcPoolKey := fmt.Sprintf(hcipamTypes.NSPoolKey, pool.Namespace, pool.Name)
	hcPoolStr, err := json.Marshal(pool)

	if err != nil {
		return fmt.Errorf("unmarshal %s to hcPool err %v ", hcPoolStr, err)
	}
	err = etcdcli.PutKV(hcPoolKey, string(hcPoolStr))
	return err

}

// GetIPPoolByName calls GetKV to get ns IPPool info from given namespace name and IPPool name
func (client *IPPoolClient) GetIPPoolByName(nsname, poolName string) *hcipamTypes.NSIPPool {
	str := etcdcli.GetKV(fmt.Sprintf(hcipamTypes.NSPoolKey, nsname, poolName))
	if str == "" {
		log.IPPool.Warnf("can not find ip pool %s", poolName)
		return nil
	}
	pool := &hcipamTypes.NSIPPool{}
	err := json.Unmarshal([]byte(str), pool)
	if err != nil {
		log.IPPool.Errorf("unmarshal %s to pool err ", str)
		return nil
	}
	return pool
}

// GetNSAllPool calls GetList to get all nsIPPools in the given namespace
func (client *IPPoolClient) GetNSAllPool(namespace string) []*hcipamTypes.NSIPPool {
	var pooList []*hcipamTypes.NSIPPool
	strList := etcdcli.GetList(fmt.Sprintf(hcipamTypes.NSPoolKey, namespace, ""))
	for _, str := range strList {
		if str == "" {
			log.IPPool.Error("empty str to unmarshal")
			continue
		}
		pool := &hcipamTypes.NSIPPool{}
		err := json.Unmarshal([]byte(str), pool)
		if err != nil {
			log.IPPool.Errorf("unmarshal %s to pool err ", str)
			return nil
		}
		pooList = append(pooList, pool)
	}
	return pooList
}

// GetServicePool calls GetKV to get service IPPool from the given namespace and service IPPool name
func (client *IPPoolClient) GetServicePool(namespace string, ServicePoolName string) (*hcipamTypes.ServiceIPPool, error) {
	str := etcdcli.GetKV(fmt.Sprintf(hcipamTypes.ServiccePoolKey, namespace, ServicePoolName))
	if str == "" {
		return nil, fmt.Errorf("can not find ServiceIPPool %s  : %s", namespace, ServicePoolName)
	}
	serviceIPPool := &hcipamTypes.ServiceIPPool{}
	err := json.Unmarshal([]byte(str), serviceIPPool)
	if err != nil {
		log.IPPool.Errorf("unmarshal %s to serviceIPPool err ", str)
		return nil, err
	}

	return serviceIPPool, nil
}

// GetAllPool calls GetList to get all ns IPPools
func (client *IPPoolClient) GetAllPool() []*hcipamTypes.NSIPPool {
	var pooList []*hcipamTypes.NSIPPool
	strList := etcdcli.GetList(hcipamTypes.NSPoolKeyPrefix)
	for _, str := range strList {
		if str == "" {
			log.IPPool.Error("empty str to unmarshal")
			continue
		}
		pool := &hcipamTypes.NSIPPool{}
		err := json.Unmarshal([]byte(str), pool)
		if err != nil {
			log.IPPool.Errorf("unmarshal %s to pool err ", str)
			return nil
		}
		pooList = append(pooList, pool)
	}
	return pooList
}

// Delete removes the given ns IPPool's record in etcd
func (client *IPPoolClient) Delete(pool *hcipamTypes.NSIPPool) error {
	e := client.Cli
	_, err := concurrency.NewSTMRepeatable(context.TODO(), e, func(s concurrency.STM) error {
		hcPoolKey := fmt.Sprintf(hcipamTypes.NSPoolKey, pool.Namespace, pool.Name)
		s.Del(hcPoolKey)
		return nil
	})
	return err
}

// CreateServiceIPPool writes a new service IPPool into etcd
func (client *IPPoolClient) CreateServiceIPPool(serviceIPPool *hcipamTypes.ServiceIPPool) error {

	var succeed = false
	var errMsg string
	e := client.Cli

	_, err := concurrency.NewSTMRepeatable(context.TODO(), e, func(s concurrency.STM) error {
		hcPoolKey := fmt.Sprintf(hcipamTypes.NSPoolKey, serviceIPPool.Namespace, serviceIPPool.NSIPPoolName)
		hcPoolStr := s.Get(hcPoolKey)
		nsPool := &hcipamTypes.NSIPPool{}
		err := json.Unmarshal([]byte(hcPoolStr), nsPool)
		if ip.Cmp(nsPool.Start, serviceIPPool.Start) > 0 || ip.Cmp(nsPool.End, serviceIPPool.End) < 0 {
			return fmt.Errorf("service IPPool out of range ,nsPool start %s,end is %s, service Pool start %s,end pool %s",
				nsPool.Start, nsPool.End, serviceIPPool.Start, serviceIPPool.End)
		}

		if err != nil {
			return fmt.Errorf("unmarshal %s to nsPool err %v ", hcPoolStr, err)
		}
		if nsPool.Total <= nsPool.Used {
			return fmt.Errorf("nsPool %s to has no ip left  ", hcPoolStr)
		}

		svcTotal := utils.GetIPTotal(serviceIPPool.Start, serviceIPPool.End)
		if (nsPool.Total - nsPool.Used) < svcTotal {
			return fmt.Errorf("nsPool %s to has no enough ip left ,%s", hcPoolStr, serviceIPPool.Name)
		}

		podMap := nsPool.PodMap

		start := serviceIPPool.Start
		end := serviceIPPool.End
		ipAddr := start
		for ip.Cmp(ipAddr, end) < 1 {
			_, exist := podMap[ipAddr.String()]
			//service pool can not  overlap with other pool
			if exist {
				errMsg = fmt.Sprintf(" ip  %s already inuse", ipAddr.String())
				return nil
			}
			ipAddr = ip.NextIP(ipAddr)
		}
		ipAddr = start
		serviceIPTotal := 0
		for ip.Cmp(ipAddr, end) < 1 {
			serviceIPTotal++
			temPodNetWrk := hcipamTypes.PodNetwork{PodName: "", SvcIPPoolName: serviceIPPool.Name,
				Namespace: serviceIPPool.Namespace, PodIP: ipAddr, SvcName: serviceIPPool.ServiceName}
			nsPool.PodMap[ipAddr.String()] = temPodNetWrk
			ipAddr = ip.NextIP(ipAddr)
		}

		nsPool.Used = nsPool.Used + serviceIPTotal

		// ns ippool
		hcPoolStrByte, err := json.Marshal(nsPool)

		if err != nil {
			return fmt.Errorf("hcPoolS  marshal err %v", err)
		}
		s.Put(fmt.Sprintf(hcipamTypes.NSPoolKey, nsPool.Namespace, nsPool.Name), string(hcPoolStrByte))

		//service ippool
		serviceIPPool.Total = serviceIPTotal
		serviceIPPool.Used = 0
		serviceIPPool.Deleted = false
		str, err := json.Marshal(serviceIPPool)
		if err != nil {
			return err
		}
		key := fmt.Sprintf(hcipamTypes.ServiccePoolKey, serviceIPPool.Namespace, serviceIPPool.Name)
		s.Put(key, string(str))
		succeed = true
		return nil
	})
	if err != nil {
		return err
	}
	if succeed {
		return nil
	}

	return fmt.Errorf(errMsg)
}

// DeleteServiceIPPool removes the given service IPPool's record in etcd by it's name and namespace
func (client *IPPoolClient) DeleteServiceIPPool(namespace string, servicePoolName string) error {
	e := client.Cli
	_, err := concurrency.NewSTMRepeatable(context.TODO(), e, func(s concurrency.STM) error {
		servicePoolKey := fmt.Sprintf(hcipamTypes.ServiccePoolKey, namespace, servicePoolName)
		svcPoolStr := s.Get(servicePoolKey)
		serviceIPPool := &hcipamTypes.ServiceIPPool{}
		err := json.Unmarshal([]byte(svcPoolStr), serviceIPPool)
		if err != nil {
			return fmt.Errorf("unmarshal %s to svcPool err %v ", svcPoolStr, err)
		}

		hcPoolKey := fmt.Sprintf(hcipamTypes.NSPoolKey, namespace, serviceIPPool.NSIPPoolName)
		hcPoolStr := s.Get(hcPoolKey)
		nsPool := &hcipamTypes.NSIPPool{}
		err = json.Unmarshal([]byte(hcPoolStr), nsPool)
		if err != nil {
			return fmt.Errorf("unmarshal %s to nsPool err %v ", hcPoolStr, err)
		}
		podMap := nsPool.PodMap
		start := serviceIPPool.Start
		end := serviceIPPool.End
		ipAddr := start
		for ip.Cmp(ipAddr, end) < 1 {
			podNetWork, exist := podMap[ipAddr.String()]
			if exist && podNetWork.PodName != "" {
				return fmt.Errorf("ip %s is inuse by pod %s", ipAddr.String(), podNetWork.PodName)
			}
			ipAddr = ip.NextIP(ipAddr)
		}
		ipAddr = start
		serviceIPTotal := 0

		for ip.Cmp(ipAddr, end) < 1 {
			serviceIPTotal++
			delete(nsPool.PodMap, ipAddr.String())
			ipAddr = ip.NextIP(ipAddr)
		}

		nsPool.Used = nsPool.Used - serviceIPTotal
		//update ns ippool
		hcPoolStrByte, err := json.Marshal(nsPool)

		if err != nil {
			return fmt.Errorf("hcPoolS  marshal err %v", err)
		}
		s.Put(fmt.Sprintf(hcipamTypes.NSPoolKey, nsPool.Namespace, nsPool.Name), string(hcPoolStrByte))

		//delete service ippool

		key := fmt.Sprintf(hcipamTypes.ServiccePoolKey, serviceIPPool.Namespace, serviceIPPool.Name)
		s.Del(key)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
