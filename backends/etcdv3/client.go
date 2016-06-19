package etcdv3

import (
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
)

// Client is a wrapper around the etcd client
type Client struct {
	client *clientv3.Client
}

// NewEtcdClient returns an *etcdv3.Client with a connection to named machines.
func NewEtcdClient(machines []string, cert, key, caCert string, basicAuth bool, username string, password string) (*Client, error) {
	var cli *clientv3.Client
	cfg := clientv3.Config{
		Endpoints:   machines,
		DialTimeout: 5 * time.Second,
	}
	tlsInfo := &transport.TLSInfo{}

	if basicAuth {
		cfg.Username = username
		cfg.Password = password
	}

	tls := false
	if caCert != "" {
		tlsInfo.CAFile = caCert
		tls = true
	}
	if cert != "" && key != "" {
		tlsInfo.CertFile = cert
		tlsInfo.KeyFile = key
		tls = true
	}

	if tls {
		clientConf, err := tlsInfo.ClientConfig()
		if err != nil {
			return &Client{cli}, err
		}
		cfg.TLS = clientConf
	}

	cli, err := clientv3.New(cfg)
	if err != nil {
		return &Client{cli}, err
	}
	return &Client{cli}, nil
}

// GetValues queries etcd for keys prefixed by prefix.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
		resp, err := c.client.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
		cancel()
		if err != nil {
			return vars, err
		}
		for _, ev := range resp.Kvs {
			vars[string(ev.Key)] = string(ev.Value)
		}
	}
	return vars, nil
}

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cancelRoutine := make(chan bool)
	defer close(cancelRoutine)
	var err error

	go func() {
		select {
		case <-stopChan:
			cancel()
		case <-cancelRoutine:
			return
		}
	}()

	rch := c.client.Watch(ctx, prefix, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			// Only return if we have a key prefix we care about.
			// This is not an exact match on the key so there is a chance
			// we will still pickup on false positives. The net win here
			// is reducing the scope of keys that can trigger updates.
			for _, k := range keys {
				if strings.HasPrefix(string(ev.Kv.Key), k) {
					return uint64(ev.Kv.Version), err
				}
			}
		}
	}
	return 0, err
}
