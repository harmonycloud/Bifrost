// Copyright 2016 CNI authors
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

package main

import (
	"fmt"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/satori/go.uuid"
	"testing"
)

func TestAdd(t *testing.T) {
	const ifname string = "eth0"
	const nspath string = "some/where"

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
		ContainerID: fmt.Sprintf("pod-%s", randomID.String()),
		Netns:       nspath,
		IfName:      ifname,
		StdinData:   []byte(conf),
	}
	cmdAdd(args)

}

func TestCmdDel(t *testing.T) {
	fmt.Println("sss")
	const ifname string = "eth0"
	const nspath string = "some/where"

	conf := fmt.Sprintf(`{
		"cniVersion": "0.3.1",
		"name": "mynet",
		"type": "ipvlan",
		"master": "foo0",
		"ipam": {
	   		"type": "hcipam",
        	"etcd_endpoints": "https://10.10.101.203:2379",
        	"etcd_key_file": "/etc/kubernetes/pki/apiserver-etcd-client.key",
        	"etcd_cert_file": "/etc/kubernetes/pki/apiserver-etcd-client.crt",
        	"etcd_ca_cert_file": "/etc/kubernetes/pki/etcd/ca.crt"
				}
		}`)

	args := &skel.CmdArgs{
		ContainerID: fmt.Sprintf("pod-%s", "0358ca92-3357-11e9-a346-8cec4b269a99"),
		Netns:       nspath,
		IfName:      ifname,
		StdinData:   []byte(conf),
	}

	cmdDel(args)

}
