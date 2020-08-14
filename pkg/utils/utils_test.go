package utils

import (
	"fmt"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	hcipamtypes "github.com/harmonycloud/bifrost/pkg/types"
	uuid "github.com/satori/go.uuid"
	"net"
	"testing"
)

func TestAdd(t *testing.T) {

	ipv4, err := types.ParseCIDR("1.2.3.30/24")

	routegwv4, routev4, err := net.ParseCIDR("15.5.6.8/24")

	const ifname string = "veth1"

	const nspath string = "/var/run/netns/ns0"

	conf := fmt.Sprintf(`
       {
            "type": "hcmacvlan",
             "ipam": {
                "type": "hcipam",
                "resolvConf": "/etc/resolv.conf",
                "log_path": "/var/log/hcipam.log",
                "log_level": "DEBUG",
                "etcd_endpoints": "https://10.10.101.203:2379",
                "etcd_key_file": "/etc/cni/net.d/ssl/etcd-key",
                "etcd_cert_file": "/etc/cni/net.d/ssl/etcd-cert",
                "etcd_ca_cert_file": "/etc/cni/net.d/ssl/etcd-ca"
             }}
`)

	randomID, _ := uuid.NewV1()

	args := &skel.CmdArgs{
		ContainerID: fmt.Sprintf("pod%s", randomID.String()),
		Netns:       nspath,
		IfName:      ifname,
		StdinData:   []byte(conf),
	}

	_, ipamConf, _, err := hcipamtypes.LoadIPAMConfig(args.StdinData, args.Args)

	if err != nil {
		fmt.Println(err)
	}
	result := &current.Result{
		CNIVersion: "0.3.1",
		IPs: []*current.IPConfig{
			{
				Version:   "4",
				Interface: current.Int(0),
				Address:   *ipv4,
				Gateway:   net.ParseIP("1.2.3.1"),
			},
		},
		Routes: []*types.Route{
			{Dst: *routev4, GW: routegwv4},
		},
	}

	clusterIPCIRDR, _ := types.ParseCIDR("1.96.0.0/24")
	hostnet, _ := types.ParseCIDR("10.100.100.102/32")
	var routerList []*net.IPNet
	routerList = append(routerList, clusterIPCIRDR)
	routerList = append(routerList, hostnet)

	DoNetworking(args, *ipamConf, result, "")
}

func TestAdd2(t *testing.T) {
	if mac, err := net.ParseMAC("EE:EE:EE:EE:EE:EE"); err != nil {
		fmt.Println("mac")
	} else {
		fmt.Println(mac)
	}

}
