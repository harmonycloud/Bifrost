package utils

import (
	"github.com/vishvananda/netlink"
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
