/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package etcdv3

import (
	"strings"
	"time"

	"context"

	"github.com/HeavyHorst/easyKV"
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

// Close closes the client connection
func (c *Client) Close() {
	c.client.Close()
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

// WatchPrefix watches a specific prefix for changes.
func (c *Client) WatchPrefix(ctx context.Context, prefix string, opts ...easyKV.WatchOption) (uint64, error) {
	var options easyKV.WatchOptions
	for _, o := range opts {
		o(&options)
	}

	etcdctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var err error

	rch := c.client.Watch(etcdctx, prefix, clientv3.WithPrefix())
	for wresp := range rch {
		if wresp.Err() != nil {
			return options.WaitIndex, wresp.Err()
		}
		for _, ev := range wresp.Events {
			// Only return if we have a key prefix we care about.
			// This is not an exact match on the key so there is a chance
			// we will still pickup on false positives. The net win here
			// is reducing the scope of keys that can trigger updates.
			for _, k := range options.Keys {
				if strings.HasPrefix(string(ev.Kv.Key), k) {
					return uint64(ev.Kv.Version), err
				}
			}
		}
	}
	if ctx.Err() == context.Canceled {
		return options.WaitIndex, easyKV.ErrWatchCanceled
	}
	return 0, err
}
