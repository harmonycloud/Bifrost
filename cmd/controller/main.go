package main

import (
	"flag"
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/harmonycloud/bifrost/pkg/allocator"
	"github.com/harmonycloud/bifrost/pkg/controller"
	"github.com/harmonycloud/bifrost/pkg/log"
	"github.com/harmonycloud/bifrost/pkg/types"
	"k8s.io/client-go/kubernetes"
	gorest "k8s.io/client-go/rest"

	"net/http"
	"os"
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

	allocator, err := allocator.NewIPAllocator(ipamConf)
	if err != nil {
		panic(err.Error())
	}
	defer allocator.Close()

	config, err := gorest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	retriever := controller.NewRetriever(allocator, clientset)

	log.Rest.Infof("cronjob %s", ipamConf.Cronjob)
	go controller.Cronjob(ipamConf.Cronjob, retriever)

	//retriever.RetrieveFixedIP()
	//go controller.Cronjob("0 0 23 L * ?", retriever)

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	server := controller.HttpServer{*allocator}

	router, err := rest.MakeRouter(
		rest.Post("/nsippool", server.CreateIPPool),
		rest.Get("/nsippool/:namespace/:poolName", server.GetIPPool),
		rest.Get("/getAllPool/:namespace", server.GetNSAllIPPool),
		rest.Get("/getAllPool", server.GetAllIPPool),
		rest.Delete("/nsippool/:namespace/:poolName", server.DeleteIPPool),
		rest.Post("/serviceIPPool", server.CreateServiceIPPool),
		rest.Get("/serviceIPPool/:namespace/:servicePoolName", server.GetServiceIPPool),
		rest.Delete("/serviceIPPool/:namespace/:servicePoolName", server.DeleteServiceIPPool),
		rest.Post("/serviceIPPool/createMultiServiceIPPool", server.CreateMultiServiceIPPool),
		rest.Delete("/:namespace/:serviceName", server.DeleteService),
		rest.Get("/:namespace/:serviceName", server.GetService),
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
