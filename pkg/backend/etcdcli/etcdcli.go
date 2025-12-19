package etcdcli

import (
	"fmt"
	"github.com/harmonycloud/bifrost/pkg/types"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/pkg/transport"
	"strings"
	"time"
)

const (
	clientTimeout    = 10 * time.Second
	keepaliveTime    = 30 * time.Second
	keepaliveTimeout = 10 * time.Second
	timeOut          = 1 * time.Second
)

type EtcdClient struct {
	clientv3.Client
}

// Init returns an etcd client from given config
func Init(config *types.IPAMConfig) (*EtcdClient, error) {
	etcdLocation := []string{}
	// Get etcd endpoints from IPAM config
	if config.EtcdEndpoints != "" {
		etcdLocation = strings.Split(config.EtcdEndpoints, ",")
	}

	if len(etcdLocation) == 0 {
		log.Warnf("No etcd endpoints specified in etcdv3 API config")
		return nil, fmt.Errorf("no etcd endpoints specified")
	}

	// Build the etcdv3 config.
	cfg := clientv3.Config{
		Endpoints:            etcdLocation,
		DialTimeout:          clientTimeout,
		DialKeepAliveTime:    keepaliveTime,
		DialKeepAliveTimeout: keepaliveTimeout,
	}

	if config.EtcdCaCertFile != "" && config.EtcdKeyFile != "" {
		// Create the etcd client
		tlsInfo := &transport.TLSInfo{
			TrustedCAFile: config.EtcdCaCertFile,
			CertFile:      config.EtcdCertFile,
			KeyFile:       config.EtcdKeyFile,
		}

		tls, err := tlsInfo.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("could not initialize etcdv3 client: %+v", err)
		}
		cfg.TLS = tls
	}

	// Plumb through the username and password if both are configured.
	if config.EtcdUsername != "" && config.EtcdPassword != "" {
		cfg.Username = config.EtcdUsername
		cfg.Password = config.EtcdPassword
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}

	return &EtcdClient{*client}, nil

}

// Close ends the connection between etcd client and etcd
func (etcdCLient EtcdClient) Close() {
	etcdCLient.Close()
}
