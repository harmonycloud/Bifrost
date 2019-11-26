# Getting started

### Prerequisite

* Linux box with
  * We tested on Linux kernel 3.10 & 4.19
* Docker installed
* Kubernetes cluster running with CNI enabled
      
### Create L2 Bridge 
```shell
nohup sh build_br0 eth0 
```

*notice:* The VLAN parameter configures the VLAN tag on the host end of the veth and also enables the vlan_filtering feature on the bridge interface.
### Installing HC-Bridge components

We install HC-bridge as a Docker Container on every node
```bash
nohup sh install.sh https://etcd_host:2379
```

### Network configuration reference
```json
{
    "cniVersion": "0.3.0",
    "name": "mynet",
    "plugins": [
        {
            "type": "hcbridge",
            "bridge": "br0",
             "ipam": {
                "type": "hcipam",
                "log_path": "/var/log/hcipam.log",
                "log_level": "DEBUG",
                "etcd_endpoints": "https://10.100.100.70:2379",
                "etcd_key_file": "/etc/cni/net.d/ssl/etcd-key",
                "etcd_cert_file": "/etc/cni/net.d/ssl/etcd-cert",
                "etcd_ca_cert_file": "/etc/cni/net.d/ssl/etcd-ca"
             }

        }
    ]
}
```
* `name` (string, required): the name of the network.
* `type` (string, required): "hcbridge".
* `bridge` (string, optional): name of the bridge to use/create. Defaults to "cni0".
* `ipam` (dictionary, required): IPAM configuration to be used for this network. For L2-only network, create empty dictionary.

