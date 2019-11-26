package main

import (
	"fmt"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/j-keck/arping"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigs := make(chan os.Signal, 1)
	//done := make(chan bool, 1)
	signal.Notify(
		sigs,
		syscall.SIGUSR1,
	)
	for {
		sig := <-sigs
		go func() {
			fmt.Println("sig:", sig)
			listen()
		}()
	}
}

const dockerNetns = "/var/run/docker/netns/"
const eth0 = "eth0"

func listen() {
	netnsList, err := ioutil.ReadDir(dockerNetns)
	if err != nil {
		log.Println(err)
		return
	}
	for _, netnsstr := range netnsList {

		fmt.Println("netns name=" + netnsstr.Name())
		netns, err := ns.GetNS(dockerNetns + netnsstr.Name())

		if err != nil {
			log.Printf("failed to open netns %q: %v", netnsstr, err)
			return
		}
		fmt.Println("netns" + netns.Path())

		err = netns.Do(func(netNS ns.NetNS) error {
			contVeth, err := net.InterfaceByName(eth0)

			if err != nil {
				return err
			}
			fmt.Println("conVeth=" + contVeth.Name)
			addrs, _ := contVeth.Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						fmt.Println(ipnet.IP.String())
						_ = arping.GratuitousArpOverIface(ipnet.IP.To4(), *contVeth)

					}
				}

			}
			return nil
		})
		if err != nil {
			log.Printf("send arp err %q: %v", netnsstr, err)
		}
		netns.Close()
	}
}
