package etcdcli

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/harmonycloud/hcbridge/pkg/log"
	"github.com/harmonycloud/hcbridge/pkg/types"
	"go.etcd.io/etcd/clientv3"
	"strings"
	"time"
)

const (
	clientTimeout    = 10 * time.Second
	keepaliveTime    = 30 * time.Second
	keepaliveTimeout = 10 * time.Second
	timeOut          = 1 * time.Second
)

var cli *clientv3.Client

// Init returns an etcd client from given config
func Init(config *types.IPAMConfig) (*clientv3.Client, error) {
	etcdLocation := []string{}
	// Get etcd endpoints from IPAM config
	if config.EtcdEndpoints != "" {
		etcdLocation = strings.Split(config.EtcdEndpoints, ",")
	}

	if len(etcdLocation) == 0 {
		log.Etcd.Warnf("No etcd endpoints specified in etcdv3 API config")
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
	cli = client
	if err != nil {
		return nil, err
	}

	return client, nil

}

// GetKV returns value for the given key
func GetKV(key string) string {
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	resp, err := cli.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		log.Etcd.Error(err)
		return ""
	}
	if resp.Count == 0 {
		return ""
	}
	return string(resp.Kvs[0].Value)
}

// GetList returns a list of values for the given prefix key
func GetList(key string) []string {
	var result []string
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	resp, err := cli.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		log.Etcd.Error(err)
		return result
	}
	if resp.Count == 0 {
		return result
	}
	for _, v := range resp.Kvs {
		result = append(result, string(v.Value))
	}
	return result
}

// PutKV writes the given KV pairs into etcd
func PutKV(key, val string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()

	_, err := cli.Put(ctx, key, val)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// Close ends the connection between etcd client and etcd
func Close() {
	cli.Close()
}
