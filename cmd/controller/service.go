package main

import (
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/harmonycloud/bifrost/pkg/etcdcli"
	"github.com/harmonycloud/bifrost/pkg/ippool"
	"github.com/harmonycloud/bifrost/pkg/types"
	"log"
	"net"
	"net/http"
)

// CreateServiceIPPool parses the input config and create a new serviceIPPool in etcd
func CreateServiceIPPool(w rest.ResponseWriter, r *rest.Request) {
	lock.Lock()
	defer lock.Unlock()
	servicePool := &types.ServiceIPPool{}
	err := r.DecodeJsonPayload(&servicePool)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if servicePool.NSIPPoolName == "" {
		rest.Error(w, " NSIPPoolName Name required", 400)
		return
	}
	if servicePool.Name == "" {
		rest.Error(w, " service pool name required", 400)
		return
	}

	if servicePool.Namespace == "" {
		rest.Error(w, " Namespace required", 400)
		return
	}
	if servicePool.ServiceName == "" {
		rest.Error(w, " ServiceName required", 400)
		return
	}
	etcdclient, err := etcdcli.Init(ipamConf)
	ippClient := ippool.NewIPPoolClient(etcdclient)
	defer etcdcli.Close()
	hcPool := ippClient.GetIPPoolByName(servicePool.Namespace, servicePool.NSIPPoolName)
	if hcPool == nil {
		rest.Error(w, fmt.Sprintf(" can not find ippoolcli %s", servicePool.NSIPPoolName), 400)
		return
	}

	if servicePool.Start == nil || servicePool.Start.Equal(net.IPv4(0, 0, 0, 0)) {
		rest.Error(w, " CIDR or start ip required", 400)
		return
	}
	if servicePool.End == nil || servicePool.End.Equal(net.IPv4(0, 0, 0, 0)) {
		rest.Error(w, " CIDR or end ip ip required", 400)
		return
	}
	validate := &types.NSIPPool{
		Start:   servicePool.Start,
		End:     servicePool.End,
		Subnet:  hcPool.Subnet,
		Gateway: hcPool.Gateway,
	}
	err = validate.Canonicalize()
	if err != nil {
		rest.Error(w, fmt.Sprintf("pool Canonicalize err %v", err), 400)
		return
	}
	err = ippClient.CreateServiceIPPool(servicePool)
	if err != nil {
		rest.Error(w, fmt.Sprintf("pool Canonicalize err %v", err), 400)
		return
	}
	err = w.WriteJson(&servicePool)
	if err != nil {
		log.Fatal(err)
	}
}

// GetServiceIPPool returns a serviceIPPool by the given name and namespace
func GetServiceIPPool(w rest.ResponseWriter, r *rest.Request) {
	lock.Lock()
	defer lock.Unlock()

	namespace := r.PathParam("namespace")
	servicePoolName := r.PathParam("servicePoolName")
	if namespace == "" {
		rest.Error(w, " Namespace required", 400)
		return
	}
	if servicePoolName == "" {
		rest.Error(w, " servicePoolName required", 400)
		return
	}
	etcdclient, err := etcdcli.Init(ipamConf)
	ippClient := ippool.NewIPPoolClient(etcdclient)
	if err != nil {
		rest.Error(w, fmt.Sprintf("system err %v", err), 500)
		return
	}
	defer etcdcli.Close()
	servicePool, err := ippClient.GetServicePool(namespace, servicePoolName)
	if err != nil {
		rest.Error(w, fmt.Sprintf("GetServicePool err %v", err), 400)
		return
	}
	err = w.WriteJson(&servicePool)
	if err != nil {
		log.Fatal(err)
	}

}

// DeleteServiceIPPool deletes a serviceIPPool in etcd by the given name and namespace
func DeleteServiceIPPool(w rest.ResponseWriter, r *rest.Request) {
	lock.Lock()
	defer lock.Unlock()

	namespace := r.PathParam("namespace")
	servicePoolName := r.PathParam("servicePoolName")
	if namespace == "" {
		rest.Error(w, " Namespace required", 400)
		return
	}
	if servicePoolName == "" {
		rest.Error(w, " servicePoolName required", 400)
		return
	}
	etcdclient, err := etcdcli.Init(ipamConf)
	ippClient := ippool.NewIPPoolClient(etcdclient)
	if err != nil {
		rest.Error(w, fmt.Sprintf("system err %v", err), 500)
		return
	}
	servicePool, err := ippClient.GetServicePool(namespace, servicePoolName)

	if err != nil {
		rest.Error(w, fmt.Sprintf("can not find ippool  %s", servicePoolName), 400)
		return
	}
	err = ippClient.DeleteServiceIPPool(namespace, servicePoolName)
	if err != nil {
		rest.Error(w, fmt.Sprintf("system err %v", err), 500)
		return
	}
	err = w.WriteJson(&servicePool)
	if err != nil {
		log.Fatal(err)
	}
}
