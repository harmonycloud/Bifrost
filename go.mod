module github.com/harmonycloud/bifrost

go 1.13

require (
	github.com/ant0ine/go-json-rest v3.3.3-0.20170913041208-ebb33769ae01+incompatible
	github.com/antonfisher/nested-logrus-formatter v1.0.2
	github.com/containernetworking/cni v0.7.2-0.20190703150356-dc71cd2ba60c
	github.com/containernetworking/plugins v0.8.4-0.20191113162045-497560f35f2c
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/go-logr/logr v0.1.1-0.20190903151443-a1ebd699b195 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/googleapis/gnostic v0.4.0
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.9.6 // indirect
	github.com/harmonycloud/hcbridge v0.0.0-20200810022910-c4e93b702089
	github.com/imdario/mergo v0.3.10 // indirect
	github.com/j-keck/arping v1.0.0
	github.com/jonboulle/clockwork v0.2.0 // indirect
	github.com/onsi/gomega v1.7.0
	github.com/prometheus/client_golang v1.7.1 // indirect
	github.com/robfig/cron v1.2.0
	github.com/satori/go.uuid v1.2.1-0.20180103174451-36e9d2ebbde5
	github.com/sirupsen/logrus v1.4.2
	github.com/tmc/grpc-websocket-proxy v0.0.0-20200427203606-3cfed13b9966 // indirect
	github.com/vishvananda/netlink v1.1.0
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200401174654-e694b7bb0875
	go.uber.org/zap v1.15.0 // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	google.golang.org/grpc v1.26.0
	k8s.io/api v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/apiserver v0.17.3 // indirect
	k8s.io/client-go v0.17.3
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20200720150651-0bdb4ca86cbc // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
