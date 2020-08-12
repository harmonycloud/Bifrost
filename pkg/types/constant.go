package types

// etcd key templates or prefixes,
// pod annotation keys
const (
	NS_POOL_Key = "ipam.harmonycloud.cn/nspool/%s/%s" //poolname

	NS_POOL_Key_Prefix = "ipam.harmonycloud.cn/nspool"

	Servicce_Pool_Key = "ipam.harmonycloud.cn/svcpool/%s/%s" //poolname

	Servicce_Pool_Key_Prefix = "ipam.harmonycloud.cn/svcpool"

	Pod_Key = "ipam.harmonycloud.cn/pod/%s/%s" //ns podname

	Pod_Key_Prefix = "ipam.harmonycloud.cn/pod/"

	defaultIpPool = "default_ippool"

	POD_IP_FIXED = "hcipm_ip_fixed"

	POD_IP_ANNOTAION = "hcipam_ip"

	POD_IP_POOL = "hcipam_ns_ippool"

	SERVICE_IP_POOL = "hcipam_service_ippool"

	Service_Key_Prefix = "hcipam_service"

	Service_Key = "ipam.harmonycloud.cn/hcipam_service/%s/%s"
)
