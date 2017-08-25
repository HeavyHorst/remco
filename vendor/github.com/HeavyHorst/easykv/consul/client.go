/*
 * This file is part of easyKV.
 * Based on code from confd.
 * https://github.com/kelseyhightower/confd/blob/2cacfab234a5d61be4cd88b9e97bee44437c318d/backends/consul/client.go
 * Users who have contributed to this file
 * © 2013 Kelsey Hightower
 * © 2015 Philip Southam
 *
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package consul

import (
	"context"
	"path"
	"strings"

	"github.com/HeavyHorst/easykv"
	"github.com/hashicorp/consul/api"
)

// Client is a wrapper around the consul KV-client.
type Client struct {
	client *api.KV
}

// New returns a new client to Consul for the given address.
func New(nodes []string, opts ...Option) (*Client, error) {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	conf := api.DefaultConfig()

	conf.Scheme = options.Scheme

	if len(nodes) > 0 {
		conf.Address = nodes[0]
	}

	tlsConfig := api.TLSConfig{}
	if options.TLS.ClientCert != "" && options.TLS.ClientKey != "" {
		tlsConfig.CertFile = options.TLS.ClientCert
		tlsConfig.KeyFile = options.TLS.ClientKey
	}

	if options.TLS.ClientCaKeys != "" {
		tlsConfig.CAFile = options.TLS.ClientCaKeys
	}

	conf.TLSConfig = tlsConfig

	client, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}
	return &Client{client.KV()}, nil
}

// Close is only meant to fulfill the easykv.ReadWatcher interface.
// Does nothing.
func (c *Client) Close() {}

// GetValues is used to lookup all keys with a prefix.
// Several prefixes can be specified in the keys array.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		key := strings.TrimPrefix(key, "/")
		pairs, _, err := c.client.List(key, nil)
		if err != nil {
			return vars, err
		}
		for _, p := range pairs {
			vars[path.Join("/", p.Key)] = string(p.Value)
		}
	}
	return vars, nil
}

type watchResponse struct {
	waitIndex uint64
	err       error
}

// WatchPrefix watches a specific prefix for changes.
func (c *Client) WatchPrefix(ctx context.Context, prefix string, opts ...easykv.WatchOption) (uint64, error) {
	var options easykv.WatchOptions
	for _, o := range opts {
		o(&options)
	}

	respChan := make(chan watchResponse)
	go func() {
		opts := api.QueryOptions{
			WaitIndex: options.WaitIndex,
		}
		_, meta, err := c.client.List(prefix, &opts)
		if err != nil {
			respChan <- watchResponse{options.WaitIndex, err}
			return
		}
		respChan <- watchResponse{meta.LastIndex, err}
	}()
	for {
		select {
		case <-ctx.Done():
			return options.WaitIndex, easykv.ErrWatchCanceled
		case r := <-respChan:
			return r.waitIndex, r.err
		}
	}
}
