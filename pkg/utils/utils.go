package utils

import (
	"github.com/harmonycloud/bifrost/pkg/types"
	"net"
	"strconv"
	"strings"
)

// GetTotal returns the number of IP addresses in the given ns IPPool
func GetTotal(pool types.NSIPPool) int {
	startIP := stringIPToInt(pool.Start.String())
	endIP := stringIPToInt(pool.End.String())
	return endIP - startIP + 1
}

// GetIPTotal returns the number of IP addresses for the given start and end IP
func GetIPTotal(start net.IP, end net.IP) int {
	startIP := stringIPToInt(start.String())
	endIP := stringIPToInt(end.String())
	return endIP - startIP + 1
}

func stringIPToInt(ipstring string) int {
	ipSegs := strings.Split(ipstring, ".")
	var ipInt = 0
	var pos uint = 24
	for _, ipSeg := range ipSegs {
		tempInt, _ := strconv.Atoi(ipSeg)
		tempInt = tempInt << pos
		ipInt = ipInt | tempInt
		pos -= 8
	}
	return ipInt
}
