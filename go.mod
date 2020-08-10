module github.com/harmonycloud/bifrost

go 1.13

require (
	github.com/alecthomas/template v0.0.0-20160405071501-a0175ee3bccc
	github.com/alecthomas/units v0.0.0-20151022065526-2efee857e7cf
	github.com/ant0ine/go-json-rest v3.3.3-0.20170913041208-ebb33769ae01+incompatible
	github.com/antonfisher/nested-logrus-formatter v1.0.2
	github.com/containernetworking/cni v0.7.2-0.20190703150356-dc71cd2ba60c
	github.com/containernetworking/plugins v0.8.4-0.20191113162045-497560f35f2c
	github.com/coreos/etcd v3.3.18+incompatible
	github.com/go-logr/logr v0.1.1-0.20190903151443-a1ebd699b195
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/gofuzz v0.0.0-20170612174753-24818f796faf
	github.com/harmonycloud/hcbridge v0.0.0-20200810022910-c4e93b702089
	github.com/j-keck/arping v1.0.0
	github.com/prometheus/common v0.0.0-20180801064454-c7de2306084e
	github.com/satori/go.uuid v1.2.1-0.20180103174451-36e9d2ebbde5
	github.com/sirupsen/logrus v1.0.7-0.20180827052211-51df1d314861
	github.com/vishvananda/netlink v1.1.0
	github.com/vishvananda/netns v0.0.0-20191106174202-0a2b9b5464df
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.0.0-20180828152522-d150a5833232
	k8s.io/apimachinery v0.0.0-20180828123425-c6b66c9c507a
	k8s.io/client-go v0.0.0-20180828232743-87935b98dd4a
)

replace go.etcd.io/etcd v3.3.0-rc.0.0.20180829024427-e8b940f268a8+incompatible => github.com/etcd-io/etcd v0.5.0-alpha.5.0.20191028151140-84e2788c2e41

replace github.com/coreos/etcd v3.3.0-rc.0.0.20180829024427-e8b940f268a8+incompatible => github.com/etcd-io/etcd v0.5.0-alpha.5.0.20191028151140-84e2788c2e41
