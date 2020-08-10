package main

import (
	"fmt"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/j-keck/arping"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"syscall"
	"unsafe"
)

func main() {
	GetNotifyArp("bond0") // test
	//sigs := make(chan os.Signal, 1)
	////done := make(chan bool, 1)
	//signal.Notify(
	//	sigs,
	//	syscall.SIGUSR1,
	//	syscall.Signal(syscall.RTM_F_NOTIFY),
	//)
	//for {
	//	sig := <-sigs
	//	go func() {
	//		fmt.Println("sig:", sig)
	//		listen()
	//	}()
	//}
}

func GetNotifyArp(bond string) {
	l, _ := ListenNetlink()

	for {
		msgs, err := l.ReadMsgs()
		if err != nil {
			fmt.Printf("Could not read netlink:\n %s", err) // can't find this netlink
		}
	loop:
		for _, m := range msgs {
			switch m.Header.Type {
			case syscall.NLMSG_DONE, syscall.NLMSG_ERROR:
				break loop
			case syscall.RTM_NEWLINK, syscall.RTM_DELLINK: // get netlink message
				res, err := PrintLinkMsg(&m)
				if err != nil {
					fmt.Printf("Could not find netlink %s\n", err)
				} else {
					ethInfo := strings.Fields(res)
					if ethInfo[2] == bond && ethInfo[1] == "up" {
						listen()
					}
				}
			}

		}
	}
}

type NetlinkListener struct {
	fd int
	sa *syscall.SockaddrNetlink
}

func ListenNetlink() (*NetlinkListener, error) { // Listen netlink
	groups := syscall.RTNLGRP_LINK
	//|
	//syscall.RTNLGRP_IPV4_IFADDR |
	//syscall.RTNLGRP_IPV4_ROUTE |
	//syscall.RTNLGRP_IPV6_IFADDR |
	//syscall.RTNLGRP_IPV6_ROUTE

	s, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_DGRAM,
		syscall.NETLINK_ROUTE)
	if err != nil {
		return nil, fmt.Errorf("socket: %s", err)
	}

	saddr := &syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Pid:    uint32(0),
		Groups: uint32(groups),
	}

	err = syscall.Bind(s, saddr)
	if err != nil {
		return nil, fmt.Errorf("bind: %s", err)
	}

	return &NetlinkListener{fd: s, sa: saddr}, nil
}

func (l *NetlinkListener) ReadMsgs() ([]syscall.NetlinkMessage, error) { // read netlink message
	defer func() {
		recover()
	}()

	pkt := make([]byte, 2048)

	n, err := syscall.Read(l.fd, pkt)
	if err != nil {
		return nil, fmt.Errorf("read: %s", err)
	}

	msgs, err := syscall.ParseNetlinkMessage(pkt[:n])
	if err != nil {
		return nil, fmt.Errorf("parse: %s", err)
	}

	return msgs, nil
}

func PrintLinkMsg(msg *syscall.NetlinkMessage) (string, error) { // when netlink changed, function can listen the message and notify user
	defer func() {
		recover()
	}()

	var str, res string
	ifim := (*syscall.IfInfomsg)(unsafe.Pointer(&msg.Data[0]))
	eth, err := net.InterfaceByIndex(int(ifim.Index))
	if err != nil {
		return "", fmt.Errorf("LinkDev %d: %s", int(ifim.Index), err)
	}
	if eth.Flags&syscall.IFF_UP == 1 {
		str = "up"
	} else {
		str = "down"
	}
	if msg.Header.Type == syscall.RTM_NEWLINK {
		res = "NEWLINK: " + str + " " + eth.Name
	} else {
		res = "DELLINK: " + eth.Name
	}

	return res, nil
}

const dockerNetns = "/var/run/docker/netns/"
const eth0 = "eth0"

func listen() { // Let all of dockerNetns_Dev send Gratuitous ARP
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
