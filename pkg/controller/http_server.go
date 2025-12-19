package controller

import (
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/harmonycloud/bifrost/pkg/allocator"
	bflog "github.com/harmonycloud/bifrost/pkg/log"
	"github.com/harmonycloud/bifrost/pkg/types"
	"net"
	"net/http"
	"sync"
)

type HttpServer struct {
	allocator.IPAllocator
}

var log = bflog.HttpServer

var lock = sync.RWMutex{}

// GetAllIPPool returns all nsIPPools in the given namespace
func (server *HttpServer) GetAllIPPool(w rest.ResponseWriter, r *rest.Request) {
	namespace := r.PathParam("namespace")
	if namespace == "" {
		rest.Error(w, " Namespace required", 400)
		return
	}
	lock.Lock()
	defer lock.Unlock()

	pools := server.BackendClient.GetNSAllPool(namespace)
	err := w.WriteJson(pools)
	if err != nil {
		log.Fatal(err)
	}

}

// GetIPPool returns a nsIPPool by the given name and namespace
func (server *HttpServer) GetIPPool(w rest.ResponseWriter, r *rest.Request) {
	lock.Lock()
	defer lock.Unlock()
	namespace := r.PathParam("namespace")
	poolName := r.PathParam("poolName")
	if namespace == "" {
		rest.Error(w, " Namespace required", 400)
		return
	}
	if poolName == "" {
		rest.Error(w, " poolName required", 400)
		return
	}

	existPool := server.BackendClient.GetIPPoolByName(namespace, poolName)
	err := w.WriteJson(existPool)
	if err != nil {
		log.Fatal(err)
	}

}

// CreateIPPool parses the input config and create a new nsIPPool in etcd
func (server *HttpServer) CreateIPPool(w rest.ResponseWriter, r *rest.Request) {
	lock.Lock()
	defer lock.Unlock()

	nsPool := &types.NSIPPool{}
	err := r.DecodeJsonPayload(&nsPool)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if nsPool.Namespace == "" {
		rest.Error(w, " Namespace required", 400)
		return
	}
	if nsPool.Name == "" {
		rest.Error(w, " Name required", 400)
		return
	}
	if nsPool.Subnet.IP == nil || nsPool.Subnet.IP.Equal(net.IPv4(0, 0, 0, 0)) {
		rest.Error(w, " Subnet required", 400)
		return
	}
	if nsPool.CIDR.IP == nil || nsPool.CIDR.IP.Equal(net.IPv4(0, 0, 0, 0)) {
		if nsPool.Start == nil || nsPool.Start.Equal(net.IPv4(0, 0, 0, 0)) {
			rest.Error(w, " CIDR or start ip required", 400)
			return
		}
		if nsPool.End == nil || nsPool.End.Equal(net.IPv4(0, 0, 0, 0)) {
			rest.Error(w, " CIDR or end ip ip required", 400)
			return
		}
	}
	if nsPool.Gateway == nil {
		rest.Error(w, " Gateway required", 400)
		return
	}
	if nsPool.VlanId < 0 || nsPool.VlanId > 4094 {
		rest.Error(w, " VlanId out of range", 400)
		return
	}

	err = nsPool.Canonicalize()
	if err != nil {
		rest.Error(w, fmt.Sprintf("nsPool Canonicalize err %v", err), 400)
		return
	}
	ippClient := server.BackendClient
	existPool := server.BackendClient.GetIPPoolByName(nsPool.Namespace, nsPool.Name)
	if existPool != nil {
		rest.Error(w, fmt.Sprintf("nsPool %s exist !!!", nsPool.Name), 400)
		return
	}

	pooList := ippClient.GetAllPool()
	for _, v := range pooList {
		if v.Overlaps(*nsPool) {
			rest.Error(w, fmt.Sprintf("create ippool err,nsPool %s : %s is overlaps of %s:  %s  ", nsPool.Namespace, nsPool.Name, v.Namespace, v.Name), 400)
			return
		}
	}

	err = ippClient.CreateIPPool(nsPool)
	if err != nil {
		rest.Error(w, fmt.Sprintf("create ippool err %v", err), 400)
		return
	}
	err = w.WriteJson(&nsPool)
	if err != nil {
		log.Fatal(err)
	}
}

// DeleteIPPool deletes a nsIPPool in etcd by the given name and namespace
func (server *HttpServer) DeleteIPPool(w rest.ResponseWriter, r *rest.Request) {
	lock.Lock()
	defer lock.Unlock()
	namespace := r.PathParam("namespace")
	poolName := r.PathParam("poolName")
	if namespace == "" {
		rest.Error(w, " Namespace required", 400)
		return
	}
	if poolName == "" {
		rest.Error(w, " poolName required", 400)
		return
	}
	ippClient := server.BackendClient
	pool := ippClient.GetIPPoolByName(namespace, poolName)
	if pool == nil {
		rest.Error(w, fmt.Sprintf("can not find  ippool  %s", poolName), 400)
		return
	}

	if pool.Used == 0 {
		err := ippClient.Delete(pool)
		if err != nil {
			rest.Error(w, fmt.Sprintf("delete  ippool  %s", err), 400)
			return
		}
		err = w.WriteJson(&pool)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	rest.Error(w, fmt.Sprintf("can not delete ippool %s", pool.Name), 400)
	return
}

func (server *HttpServer) GetNSAllIPPool(w rest.ResponseWriter, r *rest.Request) {
	namespace := r.PathParam("namespace")
	if namespace == "" {
		rest.Error(w, " Namespace required", 400)
		return
	}
	lock.Lock()
	defer lock.Unlock()

	ippClient := server.BackendClient
	var pools []*types.NSIPPool

	pools = ippClient.GetNSAllPool(namespace)

	err := w.WriteJson(pools)
	if err != nil {
		log.Fatal(err)
	}

}
