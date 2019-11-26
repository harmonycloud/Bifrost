package types

// etcd key templates or prefixes,
// pod annotation keys
const (
	NSPoolKey = "ipam.harmonycloud.cn/nspool/%s/%s" //ns poolname

	NSPoolKeyPrefix = "ipam.harmonycloud.cn/nspool"

	ServiccePoolKey = "ipam.harmonycloud.cn/svcpool/%s/%s" //ns poolname

	ServiccePoolKeyPrefix = "ipam.harmonycloud.cn/svcpool"

	PodKey = "ipam.harmonycloud.cn/pod/%s/%s" //ns podname

	PodKeyPrefix = "ipam.harmonycloud.cn/pod/"

	PodIPFixedAnno = "hcipm_ip_fixed"

	PodIPAnno = "hcipam_ip"

	PodIPPoolAnno = "hcipam_ns_ippool"

	ServiceIPPoolAnno = "hcipam_service_ippool"
)
