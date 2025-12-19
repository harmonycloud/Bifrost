package types

import (
	"net"
)

type CreateServiceRequest struct {
	NsPoolMap map[string][]SvcPool `json:"nsPoolMap"`
	Namespace string               `json:"namespace"`
	SvcName   string               `json:"svcName"`
}

type SvcPool struct {
	Start         net.IP `json:"start"` // The first ip, inclusive
	End           net.IP `json:"end"`   // The last ip, inclusive
	SvcIPPoolName string `json:"svcIpPoolName"`
}
