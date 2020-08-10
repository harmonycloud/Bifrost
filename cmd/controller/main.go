package main

import (
	"flag"
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/harmonycloud/bifrost/pkg/controller"
	"github.com/harmonycloud/bifrost/pkg/etcdcli"
	"github.com/harmonycloud/bifrost/pkg/ippool"
	"github.com/harmonycloud/bifrost/pkg/log"
	"github.com/harmonycloud/bifrost/pkg/types"
	"k8s.io/client-go/kubernetes"
	gorest "k8s.io/client-go/rest"

	"net"
	"net/http"
	"os"
	"sync"
)

var ipamConf *types.IPAMConfig

func init() {
	log.Init("/var/log/hcipamrest.log", "Info")
}

func getIPAMFromCMD(cmdLine *flag.FlagSet) *types.IPAMConfig {

	endpoints := os.Getenv("ETCD_ENDPOINTS")
	cronjob := os.Getenv("cronjob")
	cacert := cmdLine.String("cacert", "/etc/kubernetes/pki/etcd/ca.crt", "cacert")
	cert := cmdLine.String("cert", "/etc/kubernetes/pki/apiserver-etcd-client.crt", "cert")
	key := cmdLine.String("key", "/etc/kubernetes/pki/apiserver-etcd-client.key", "key")
	cmdLine.Parse(os.Args[1:])
	fmt.Printf("endpoints is %s,cacert is %s,cert is %s,key is %s,", endpoints, *cacert, *cert, *key)
	ipamConf := &types.IPAMConfig{
		EtcdEndpoints:  endpoints,
		EtcdKeyFile:    *key,
		EtcdCertFile:   *cert,
		EtcdCaCertFile: *cacert,
		Cronjob:        cronjob,
	}
	return ipamConf

}

func main() {
	defer log.Close()
	log.Rest.Info("start hcipam rest service !!!!")
	// Use our own FlagSet, because some libraries pollute the global one
	var cmdLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	ipamConf = getIPAMFromCMD(cmdLine)

	etcdclient, err := etcdcli.Init(ipamConf)
	defer etcdcli.Close()
	config, err := gorest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	retriever := controller.NewRetriever(etcdclient, clientset)

	log.Rest.Infof("cronjob %s", ipamConf.Cronjob)
	go controller.Cronjob(ipamConf.Cronjob, retriever)

	//retriever.RetrieveFixedIP()
	//go controller.Cronjob("0 0 23 L * ?", retriever)

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Post("/nsippool", CreateIPPool),
		rest.Get("/nsippool/:namespace/:poolName", GetIPPool),
		rest.Get("/getAllPool/:namespace", GetAllIPPool),
		rest.Delete("/nsippool/:namespace/:poolName", DeleteIPPool),
		rest.Post("/serviceIPPool", CreateServiceIPPool),
		rest.Get("/serviceIPPool/:namespace/:servicePoolName", GetServiceIPPool),
		rest.Delete("/serviceIPPool/:namespace/:servicePoolName", DeleteServiceIPPool),
	)
	if err != nil {
		fmt.Println(err)
		log.Rest.Fatal(err)
	}
	api.SetApp(router)
	err = http.ListenAndServe(":38080", api.MakeHandler())
	if err != nil {
		fmt.Println(err)
		log.Rest.Fatal(err)
	}
}

var lock = sync.RWMutex{}

// GetAllIPPool returns all nsIPPools in the given namespace
func GetAllIPPool(w rest.ResponseWriter, r *rest.Request) {
	namespace := r.PathParam("namespace")
	if namespace == "" {
		rest.Error(w, " Namespace required", 400)
		return
	}
	lock.Lock()
	defer lock.Unlock()
	etcdclient, err := etcdcli.Init(ipamConf)
	defer etcdcli.Close()
	ippClient := ippool.NewIPPoolClient(etcdclient)
	if err != nil {
		log.Rest.Fatal(err)
	}
	pools := ippClient.GetNSAllPool(namespace)
	err = w.WriteJson(pools)
	if err != nil {
		log.Rest.Fatal(err)
	}

}

// GetIPPool returns a nsIPPool by the given name and namespace
func GetIPPool(w rest.ResponseWriter, r *rest.Request) {
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

	etcdclient, err := etcdcli.Init(ipamConf)
	defer etcdcli.Close()
	if err != nil {
		log.Rest.Fatal(err)
	}
	ippClient := ippool.NewIPPoolClient(etcdclient)
	existPool := ippClient.GetIPPoolByName(namespace, poolName)
	err = w.WriteJson(existPool)
	if err != nil {
		log.Rest.Fatal(err)
	}

}

// CreateIPPool parses the input config and create a new nsIPPool in etcd
func CreateIPPool(w rest.ResponseWriter, r *rest.Request) {
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
	if nsPool.VlanID < 0 || nsPool.VlanID > 4094 {
		rest.Error(w, " VlanId out of range", 400)
		return
	}

	err = nsPool.Canonicalize()
	if err != nil {
		rest.Error(w, fmt.Sprintf("nsPool Canonicalize err %v", err), 400)
		return
	}

	etcdclient, err := etcdcli.Init(ipamConf)
	defer etcdcli.Close()

	ippClient := ippool.NewIPPoolClient(etcdclient)
	if err != nil {
		log.Rest.Fatal(err)
	}

	existPool := ippClient.GetIPPoolByName(nsPool.Namespace, nsPool.Name)
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
		log.Rest.Fatal(err)
	}
}

// DeleteIPPool deletes a nsIPPool in etcd by the given name and namespace
func DeleteIPPool(w rest.ResponseWriter, r *rest.Request) {
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
	etcdclient, err := etcdcli.Init(ipamConf)
	defer etcdcli.Close()

	ippClient := ippool.NewIPPoolClient(etcdclient)
	if err != nil {
		log.Rest.Fatal(err)
	}
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
			log.Rest.Fatal(err)
		}
		return
	}
	rest.Error(w, fmt.Sprintf("can not delete ippool %s", pool.Name), 400)
	return
}
