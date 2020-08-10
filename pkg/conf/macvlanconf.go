package conf

import "github.com/containernetworking/cni/pkg/types"

type NetConf struct {
	types.NetConf
	Master string `json:"master"`
	Mode   string `json:"mode"`
	MTU    int    `json:"mtu"`
	Vlan   int    `json:"vlan"`
	Debug  string `json:"debug"`
}
