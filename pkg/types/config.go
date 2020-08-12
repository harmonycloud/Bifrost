// Copyright 2015 CNI authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/containernetworking/cni/pkg/types"
	"time"
)

// Net is the top-level network config - IPAM plugins are passed the full configuration
// of the calling plugin, not just the IPAM section.
type Net struct {
	Name       string      `json:"name"`
	CNIVersion string      `json:"cniVersion"`
	Bridge     string      `json:"bridge"`
	IPAM       *IPAMConfig `json:"ipam"`
}

// IPAMConfig represents the IP related network configuration.
// This nests NSIPPool because we initially only supported a single
// range directly, and wish to preserve backwards compatibility
type IPAMConfig struct {
	Name           string
	Type           string `json:"type"`
	MTU            int    `json:"mtu"`
	DatastoreType  string `json:"datastore_type"`
	EtcdEndpoints  string `json:"etcd_endpoints"`
	LogPath        string `json:"log_path"`
	LogLevel       string `json:"log_level"`
	EtcdKeyFile    string `json:"etcd_key_file"`
	EtcdCertFile   string `json:"etcd_cert_file"`
	EtcdCaCertFile string `json:"etcd_ca_cert_file"`
	EtcdUsername   string `json:"etcd_username"`
	EtcdPassword   string `json:"etcd_password"`
	Cronjob        string `json:"cronjob"`
}

// NSIPPool is a IPPool in a certain namespace
type NSIPPool struct {
	Name    string      `json:"name"`
	Start   net.IP      `json:"start"` // The first ip, inclusive
	End     net.IP      `json:"end"`   // The last ip, inclusive
	Subnet  types.IPNet `json:"subnet"`
	Gateway net.IP      `json:"gateway"`

	//Routes []*types.Route `json:"routes,omitempty"`
	Total int `json:"total"`
	Used  int `json:"used"`

	CIDR types.IPNet `json:"cidr,omitempty"`

	VlanId int `json:"vlanid,omitempty"`

	Namespace string `json:"namespace"`

	LastAssignedIP net.IP `json:"lastAssignedIP,omitempty"`

	PodMap map[string]PodNetwork `json:"podMap"`
}

// PodNetwork contains all info about an IP address
type PodNetwork struct {
	PodIP         net.IP    `json:"podIP,omitempty"`
	PodName       string    `json:"podName,omitempty"`
	SvcName       string    `json:"svcName,omitempty"`
	SvcIPPoolName string    `json:"svcIPPoolName,omitempty"`
	Namespace     string    `json:"namespace,omitempty"`
	StsName       string    `json:"stsName,omitempty"`
	StsDeleted    bool      `json:"stsDeleted,omitempty"`
	FixedIP       bool      `json:"fixedIP,omitempty"`
	ModifyTime    time.Time `json:"modifyTime,omitempty"`
	NsPoolName    string    `json:"NsPoolName,omitempty"`
}

// ServiceIPPool is a IPPool for a certain service
type ServiceIPPool struct {
	Name  string `json:"name"`  //namespace-service-ippoolcli-n
	Start net.IP `json:"start"` // The first ip, inclusive
	End   net.IP `json:"end"`   // The last ip, inclusive
	//Routes []*types.Route `json:"routes,omitempty"`
	Total          int    `json:"total"`
	Used           int    `json:"used"`
	ServiceName    string `json:"serviceName"`
	Namespace      string `json:"namespace"`
	NSIPPoolName   string `json:"nsIPPoolName"`
	Deleted        bool   `json:"deleted,omitempty"`
	LastAssignedIP net.IP `json:"lastAssignedIP,omitempty"`
}

type IPAMEnvArgs struct {
	types.CommonArgs
	IP net.IP `json:"ip,omitempty"`
}

// LoadIPAMConfig loads a IPAMConfig from byte slice.
func LoadIPAMConfig(bytes []byte, envArgs string) (string, *IPAMConfig, string, error) {
	n := Net{}
	if err := json.Unmarshal(bytes, &n); err != nil {
		return "", nil, "", err
	}
	if n.IPAM == nil {
		return "", nil, "", fmt.Errorf("IPAM config missing 'ipam' key")
	}

	// Copy net name into IPAM so not to drag Net struct around
	n.IPAM.Name = n.Name

	return n.Bridge, n.IPAM, n.CNIVersion, nil
}
