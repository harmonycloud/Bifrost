// Copyright 2017 CNI authors
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
	"fmt"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/plugins/pkg/ip"
	"math/big"
	"net"
)

// Canonicalize takes a given range and ensures that all information is consistent,
// filling out Start, End, and Gateway with sane values if missing
func (ippool *NSIPPool) Canonicalize() error {

	if ip.Cmp(ippool.Gateway, ippool.Start) >= 0 && ip.Cmp(ippool.Gateway, ippool.End) <= 0 {
		return fmt.Errorf("IPPool %s  can not contain gw", (*net.IPNet)(&ippool.Subnet).String())
	}
	if err := canonicalizeIP(&ippool.Subnet.IP); err != nil {
		return err
	}

	// Check if subnet is valid
	subnetOnes, masklen := ippool.Subnet.Mask.Size()
	if subnetOnes > masklen-2 {
		return fmt.Errorf("Subnet %s too small to allocate from", (*net.IPNet)(&ippool.Subnet).String())
	}
	if len(ippool.Subnet.IP.To4()) != len(ippool.Subnet.Mask) {
		return fmt.Errorf("Subnet IPNet IP and Mask version mismatch")
	}

	networkIP := ippool.Subnet.IP.Mask(ippool.Subnet.Mask)
	if !ippool.Subnet.IP.Equal(networkIP) {
		return fmt.Errorf("Subnet has host bits set. For a subnet mask of length %d the network address is %s", subnetOnes, networkIP.String())
	}
	// Get startIP and endIP from CIDR
	if ippool.CIDR.IP != nil && !ippool.CIDR.IP.Equal(net.IPv4(0, 0, 0, 0)) {
		// Check if CIDR is valid
		cidrOnes, masklen := ippool.CIDR.Mask.Size()
		if cidrOnes > masklen-2 {
			return fmt.Errorf("CIDR %s too small to allocate from", (*net.IPNet)(&ippool.CIDR).String())
		}
		if len(ippool.CIDR.IP.To4()) != len(ippool.CIDR.Mask) {
			return fmt.Errorf("CIDR  IP and Mask version mismatch")
		}
		networkIP = ippool.CIDR.IP.Mask(ippool.CIDR.Mask)
		if !ippool.CIDR.IP.Equal(networkIP) {
			return fmt.Errorf("CIDR has host bits set. For a CIDR mask of length %d the network address is %s", subnetOnes, networkIP.String())
		}

		// Set startIP and endIP, excluding subnet-zero and the all-ones subnet
		if cidrOnes == subnetOnes {
			ippool.Start = ip.NextIP(ippool.CIDR.IP)
		} else {
			ippool.Start = ippool.CIDR.IP
		}
		ippool.End = LastIP(ippool.CIDR)

		if ippool.CIDR.IP.To4() != nil && cidrOnes == subnetOnes {
			ippool.End[3]--
		}
	} else {
		// Excluding subnet-zero and the all-ones subnet
		if ippool.Start.Equal(ippool.Subnet.IP) {
			ippool.Start = ip.NextIP(ippool.Start)
		}
		if ippool.End.Equal(LastIP(ippool.Subnet)) {
			ippool.End = ip.PrevIP(ippool.End)
		}
		// Check if startIP and endIP are valid
		startInt := InetAtoN(ippool.Start)
		endInt := InetAtoN(ippool.End)
		// Check ipv4 format
		if startInt == -1 || endInt == -1 {
			return fmt.Errorf("invalid format for start,end ip ")
		}
		// Check the size of ip pool
		if endInt-startInt+1 < 1 {
			return fmt.Errorf("ip pool is too small to allocate from. start:%s, end:%s", ippool.Start.String(), ippool.End.String())
		}
		ippool.CIDR = ippool.Subnet
	}

	// Check if startIP and endIP are in the subnet
	if !ippool.ContainsInSubnet(ippool.Start) {
		return fmt.Errorf("start ip %s not in network %s", ippool.Start.String(), (*net.IPNet)(&ippool.Subnet).String())
	}
	if !ippool.ContainsInSubnet(ippool.End) {
		return fmt.Errorf("end ip %s not in network %s", ippool.End.String(), (*net.IPNet)(&ippool.Subnet).String())
	}

	// Check if gateway is in the subnet
	if !ippool.ContainsInSubnet(ippool.Gateway) {
		return fmt.Errorf("gateway unreachable, %s not in network %s", ippool.Gateway.String(), (*net.IPNet)(&ippool.Subnet).String())
	}

	return nil
}

// ContainsInSubnet checks if a given ip is contained in the subnet of nsIPPool
func (ippool NSIPPool) ContainsInSubnet(addr net.IP) bool {
	if err := canonicalizeIP(&addr); err != nil {
		return false
	}

	subnet := (net.IPNet)(ippool.Subnet)

	// Not in network
	if !subnet.Contains(addr) {
		return false
	}

	return true
}

// Overlaps returns true if there is any overlap between ranges
func (ippool NSIPPool) Overlaps(r1 NSIPPool) bool {
	if ip.Cmp(ippool.End, r1.Start) < 0 || ip.Cmp(ippool.Start, r1.End) > 0 {
		return false
	}
	return true
}

func (ippool NSIPPool) String() string {
	return fmt.Sprintf("name %s ,subnet %s, %s-%s", ippool.Name, ippool.Subnet.IP.String(),
		ippool.Start.String(),
		ippool.End.String())
}

// canonicalizeIP makes sure a provided ip is in standard form
func canonicalizeIP(ip *net.IP) error {
	if ip.To4() != nil {
		*ip = ip.To4()
		return nil
	} else if ip.To16() != nil {
		*ip = ip.To16()
		return nil
	}
	return fmt.Errorf("IP %s not v4 nor v6", *ip)
}

// LastIP determine the last IP of a subnet, excluding the broadcast if IPv4
func LastIP(CIDR types.IPNet) net.IP {
	var end net.IP
	var netIP = CIDR.IP.To4()
	for i := 0; i < len(netIP); i++ {
		end = append(end, netIP[i]|^CIDR.Mask[i])
	}

	return end
}

// InetAtoN convert an IP address to int64
func InetAtoN(ip net.IP) int64 {
	if v := ip.To4(); v != nil {
		ret := big.NewInt(0)
		ret.SetBytes(v)
		return ret.Int64()
	}
	return -1
}
