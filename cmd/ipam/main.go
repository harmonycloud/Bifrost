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

package main

import (
	"fmt"

	"github.com/harmonycloud/bifrost/pkg/allocator"
	"github.com/harmonycloud/bifrost/pkg/cniextend"
	"github.com/harmonycloud/bifrost/pkg/etcdcli"
	"github.com/harmonycloud/bifrost/pkg/kube"
	"github.com/harmonycloud/bifrost/pkg/log"
	hcipamtypes "github.com/harmonycloud/bifrost/pkg/types"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"net"
)

func main() {
	skel.PluginMain(cmdAdd, cmdGet, cmdDel, version.All, "TODO")
}

func cmdGet(args *skel.CmdArgs) error {
	// TODO: implement
	return fmt.Errorf("not implemented")
}

func cmdAdd(args *skel.CmdArgs) error {
	// confVersion not used
	_, ipamConf, _, err := hcipamtypes.LoadIPAMConfig(args.StdinData, args.Args)

	if err != nil {
		return err
	}
	log.Init(ipamConf.LogPath, ipamConf.LogLevel)
	defer log.Close()
	log.CmdAdd.Debugf("ipamconf is %v", *ipamConf)

	etcdclient, err := etcdcli.Init(ipamConf)
	defer etcdcli.Close()
	if err != nil {
		log.CmdAdd.Errorf("pool init err is %v", err)
	}

	log.CmdAdd.Debugf("args is  %v", args.Args)
	// Initialize kubernetes arguments, including podName, podNamespace, containerID and annotations
	k8sArgs := kube.K8sArgs{}
	if err := types.LoadArgs(args.Args, &k8sArgs); err != nil {
		return err
	}
	log.CmdAdd.Debugf("k8sArgs is  %v", k8sArgs)
	nameSpace := string(k8sArgs.K8sPodNamespace)
	podName := string(k8sArgs.K8sPodName)

	result := &cniextend.VlanResult{}

	kubeClient, err := kube.NewClient()
	if err != nil {
		return err
	}
	annoMap, err := kubeClient.GetK8sPodAnnotations(nameSpace, podName)
	if err != nil {
		log.CmdAdd.Error(err)
		return err
	}
	log.CmdAdd.Debugf("kube is %v", annoMap)

	serviceIPPoolName := annoMap[hcipamtypes.SERVICE_IP_POOL]

	ipAllocator := allocator.NewIPAllocator(etcdclient)
	var ipConf *current.IPConfig
	var hcIPPool *hcipamtypes.NSIPPool
	if serviceIPPoolName != "" {
		if kubeClient.IsPodFixedIP(nameSpace, podName) {
			stsName := kubeClient.GetStsName(nameSpace, podName)
			ipConf, hcIPPool, err = ipAllocator.AllocateIP(true, nameSpace, serviceIPPoolName, podName, stsName)
		} else {
			ipConf, hcIPPool, err = ipAllocator.AllocateIP(false, nameSpace, serviceIPPoolName, podName, "")
		}
		if err != nil {
			return err
		}
	} else if annoMap[hcipamtypes.POD_IP_POOL] != "" {
		ipConf, hcIPPool, err = ipAllocator.GetIPWithoutSvcPoolPolicy(nameSpace, annoMap[hcipamtypes.POD_IP_POOL], podName)
		if err != nil || ipConf == nil {
			log.CmdAdd.Errorf("allocate ip err %s !!!", err)
			return fmt.Errorf("allocate ip err %s !!!", err)
		}
	} else {
		log.CmdAdd.Errorf("kube serviceIPPoolName required!!!")
		return fmt.Errorf("kube serviceIPPoolName required!!!")
	}

	result.IPs = append(result.IPs, ipConf)
	//lan route
	defaultDst := net.IPNet{IP: net.IPv4(0, 0, 0, 0), Mask: net.IPv4Mask(0, 0, 0, 0)}
	defaultRoute := &types.Route{Dst: defaultDst, GW: hcIPPool.Gateway}
	result.Routes = append(result.Routes, defaultRoute)
	result.VlanNum = hcIPPool.VlanId
	log.CmdAdd.Debugf("Result:\n %v", result)
	return result.Print()
}

func cmdDel(args *skel.CmdArgs) error {
	_, ipamConf, _, err := hcipamtypes.LoadIPAMConfig(args.StdinData, args.Args)
	if err != nil {
		return err
	}
	log.Init(ipamConf.LogPath, ipamConf.LogLevel)
	defer log.Close()
	k8sArgs := kube.K8sArgs{}
	if err := types.LoadArgs(args.Args, &k8sArgs); err != nil {
		return err
	}
	log.CmdDel.Debugf("ipamconf is %v", *ipamConf)

	etcdclient, err := etcdcli.Init(ipamConf)

	defer etcdcli.Close()
	if err != nil {
		log.CmdDel.Errorf("pool init err is %v", err)
	}
	podName := string(k8sArgs.K8sPodName)
	nameSpace := string(k8sArgs.K8sPodNamespace)
	alloc := allocator.NewIPAllocator(etcdclient)
	if err = alloc.Release(nameSpace, podName, false); err != nil {
		log.CmdDel.Errorf("pool init err is %v", err)
		return nil
	}

	return nil
}
