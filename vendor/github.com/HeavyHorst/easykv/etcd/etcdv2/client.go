/*
 * This file is part of easyKV.
 * Based on code from confd.
 * https://github.com/kelseyhightower/confd/blob/8d0d0efc97bfbd9e4af982eca1161ef11fc06b0b/backends/etcd/client.go
 * Users who have contributed to this file
 * © 2013 Kelsey Hightower
 * © 2014 Chris Armstrong
 * © 2015 Jens Henrik Hertz
 *
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package etcdv2

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"context"

	"github.com/HeavyHorst/easykv"
	"github.com/coreos/etcd/client"
)

// Client is a wrapper around the etcd client
type Client struct {
	client client.KeysAPI
}

// NewEtcdClient returns an *etcd.Client with a connection to named machines.
func NewEtcdClient(machines []string, cert, key, caCert string, basicAuth bool, username string, password string) (*Client, error) {
	var c client.Client
	var kapi client.KeysAPI
	var err error
	var transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}

	cfg := client.Config{
		Endpoints:               machines,
		HeaderTimeoutPerRequest: time.Duration(3) * time.Second,
	}

	if basicAuth {
		cfg.Username = username
		cfg.Password = password
	}

	if caCert != "" {
		certBytes, err := ioutil.ReadFile(caCert)
		if err != nil {
			return &Client{kapi}, err
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(certBytes)

		if ok {
			tlsConfig.RootCAs = caCertPool
		}
	}

	if cert != "" && key != "" {
		tlsCert, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return &Client{kapi}, err
		}
		tlsConfig.Certificates = []tls.Certificate{tlsCert}
	}

	transport.TLSClientConfig = tlsConfig
	cfg.Transport = transport

	c, err = client.New(cfg)
	if err != nil {
		return &Client{kapi}, err
	}

	kapi = client.NewKeysAPI(c)
	return &Client{kapi}, nil
}

// Close is only meant to fulfill the easykv.ReadWatcher interface.
// Does nothing.
func (c *Client) Close() {}

// GetValues is used to lookup all keys with a prefix.
// Several prefixes can be specified in the keys array.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		resp, err := c.client.Get(context.Background(), key, &client.GetOptions{
			Recursive: true,
			Sort:      true,
			Quorum:    true,
		})
		if err != nil {
			return vars, err
		}
		err = nodeWalk(resp.Node, vars)
		if err != nil {
			return vars, err
		}
	}
	return vars, nil
}

// nodeWalk recursively descends nodes, updating vars.
func nodeWalk(node *client.Node, vars map[string]string) error {
	if node != nil {
		key := node.Key
		if !node.Dir {
			vars[key] = node.Value
		} else {
			for _, node := range node.Nodes {
				nodeWalk(node, vars)
			}
		}
	}
	return nil
}

// WatchPrefix watches a specific prefix for changes.
func (c *Client) WatchPrefix(ctx context.Context, prefix string, opts ...easykv.WatchOption) (uint64, error) {
	var options easykv.WatchOptions
	for _, o := range opts {
		o(&options)
	}

	// Setting AfterIndex to 0 (default) means that the Watcher
	// should start watching for events starting at the current
	// index, whatever that may be.
	watcher := c.client.Watcher(prefix, &client.WatcherOptions{AfterIndex: uint64(0), Recursive: true})
	etcdctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		resp, err := watcher.Next(etcdctx)
		if err != nil {
			if err == context.Canceled {
				return options.WaitIndex, easykv.ErrWatchCanceled
			}
			switch e := err.(type) {
			case *client.Error:
				if e.Code == 401 {
					return 0, nil
				}
			}
			return options.WaitIndex, err
		}

		// Only return if we have a key prefix we care about.
		// This is not an exact match on the key so there is a chance
		// we will still pickup on false positives. The net win here
		// is reducing the scope of keys that can trigger updates.
		for _, k := range options.Keys {
			if strings.HasPrefix(resp.Node.Key, k) {
				return resp.Node.ModifiedIndex, err
			}
		}
	}
}
