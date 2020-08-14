package utils

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
)

func GetNodeMaster() string {
	rs, err := netlink.RouteListFiltered(4, nil, 0)
	if err != nil {
		return ""
	}
	for _, v := range rs {
		if v.Gw != nil {
			link, _ := netlink.LinkByIndex(v.LinkIndex)
			masterName := link.Attrs().Name
			return masterName
		}
	}
	return ""

}
func GetNodeIP() (net.Addr, error) {
	rs, err := netlink.RouteListFiltered(4, nil, 0)
	if err != nil {
		return nil, err
	}
	for _, v := range rs {
		if v.Gw != nil {
			link, _ := netlink.LinkByIndex(v.LinkIndex)
			masterName := link.Attrs().Name
			intr, err := net.InterfaceByName(masterName)
			if err != nil {
				return nil, err
			}
			addrs, err := intr.Addrs()
			return addrs[0], nil
		}
	}
	return nil, fmt.Errorf("can not found node ip %s  \n", "")

}
