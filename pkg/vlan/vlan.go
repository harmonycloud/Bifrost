package vlan

import (
	"fmt"
	"github.com/harmonycloud/bifrost/pkg/conf"

	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/vishvananda/netlink"
)

func GetVLANLink(conf *conf.NetConf, vlanid int) (netlink.Link, error) {

	ifName := fmt.Sprintf("%s-vlan-%d", conf.Master, vlanid)

	var result netlink.Link

	result, err := netlink.LinkByName(ifName)
	if err == nil && result != nil && vlanid != 0 {
		v := result.(*netlink.Vlan)
		if v != nil && v.VlanId == vlanid {

			err = netlink.LinkSetUp(result)
			if err != nil {
				return nil, fmt.Errorf("failed to set vlan %q  up: %v", ifName, err)
			}
			return result, nil
		}

	}

	maskerLink, err := netlink.LinkByName(conf.Master)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup master %q: %v", conf.Master, err)
	}
	if vlanid == 0 {
		return maskerLink, nil
	}

	// due to kernel bug we have to create with tmpname or it might
	// collide with the name on the host and error out
	tmpName, err := ip.RandomVethName()
	if err != nil {
		return nil, err
	}

	if conf.MTU <= 0 {
		conf.MTU = maskerLink.Attrs().MTU
	}

	v := &netlink.Vlan{
		LinkAttrs: netlink.LinkAttrs{
			MTU:         conf.MTU,
			Name:        tmpName,
			ParentIndex: maskerLink.Attrs().Index,
		},
		VlanId: vlanid,
	}

	if err := netlink.LinkAdd(v); err != nil {
		return nil, fmt.Errorf("failed to create vlan: %v", err)
	}

	err = ip.RenameLink(tmpName, ifName)
	if err != nil {
		return nil, fmt.Errorf("failed to rename vlan to %q: %v", ifName, err)
	}

	// Re-fetch interface to get all properties/attributes
	result, err = netlink.LinkByName(ifName)
	if err != nil {
		return nil, fmt.Errorf("failed to refetch vlan %q: %v", ifName, err)
	}
	err = netlink.LinkSetUp(result)

	if err != nil {
		return nil, fmt.Errorf("failed to set vlan %q  up: %v", ifName, err)
	}

	return result, nil

}
