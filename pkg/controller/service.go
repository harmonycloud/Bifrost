package controller

import (
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/harmonycloud/bifrost/pkg/types"
	"net"
	"net/http"
)

func (server *HttpServer) CreateMultiServiceIPPool(w rest.ResponseWriter, r *rest.Request) {
	return
}

func (server *HttpServer) CreateServiceIPPool(w rest.ResponseWriter, r *rest.Request) {
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

	ippClient := server.BackendClient

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

func (server *HttpServer) GetServiceIPPool(w rest.ResponseWriter, r *rest.Request) {
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

	ippClient := server.BackendClient
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
func (server *HttpServer) DeleteServiceIPPool(w rest.ResponseWriter, r *rest.Request) {
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

	ippClient := server.BackendClient

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

func (server *HttpServer) DeleteService(w rest.ResponseWriter, r *rest.Request) {
	lock.Lock()
	defer lock.Unlock()
	namespace := r.PathParam("namespace")
	serviceName := r.PathParam("serviceName")

	ippClient := server.BackendClient
	service := ippClient.GetService(namespace, serviceName)

	for _, v := range service.NsPoolMap {

		for _, svcPool := range v {

			servicePool, err := ippClient.GetServicePool(namespace, svcPool.SvcIPPoolName)

			if err != nil {
				rest.Error(w, fmt.Sprintf("can not find ippool  %s", svcPool.SvcIPPoolName), 400)
				return
			}
			err = ippClient.DeleteServiceIPPool(namespace, svcPool.SvcIPPoolName)
			if err != nil {
				rest.Error(w, fmt.Sprintf("system err %v", err), 500)
				return
			}
			err = w.WriteJson(&servicePool)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	//TODO Clear ip leak
	/*serviceKey := fmt.Sprintf(types.Service_Key, namespace, serviceName)
	ippClient..DeleteServiceIPPool()*/
}

func (server *HttpServer) GetService(w rest.ResponseWriter, r *rest.Request) {
	namespace := r.PathParam("namespace")
	serviceName := r.PathParam("serviceName")

	ippClient := server.BackendClient
	service := ippClient.GetService(namespace, serviceName)

	err := w.WriteJson(&service)
	if err != nil {
		log.Fatal(err)
	}

}
