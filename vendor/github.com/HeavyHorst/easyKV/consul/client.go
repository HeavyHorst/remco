/*
 * This file is part of easyKV.
 * Based on code from confd. https://github.com/kelseyhightower/confd
 * © 2013 Kelsey Hightower
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package consul

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/HeavyHorst/easyKV"
	"github.com/hashicorp/consul/api"
)

// Client provides a wrapper around the consulkv client
type Client struct {
	client *api.KV
}

// New returns a new client to Consul for the given address
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

	tlsConfig := &tls.Config{}
	if options.TLS.ClientCert != "" && options.TLS.ClientKey != "" {
		clientCert, err := tls.LoadX509KeyPair(options.TLS.ClientCert, options.TLS.ClientKey)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
		tlsConfig.BuildNameToCertificate()
	}
	if options.TLS.ClientCaKeys != "" {
		ca, err := ioutil.ReadFile(options.TLS.ClientCaKeys)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(ca)
		tlsConfig.RootCAs = caCertPool
	}
	conf.HttpClient.Transport = &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}
	return &Client{client.KV()}, nil
}

// Close closes the client connection
func (c *Client) Close() {
	return
}

// GetValues queries Consul for keys
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
func (c *Client) WatchPrefix(prefix string, stopChan chan bool, opts ...easyKV.WatchOption) (uint64, error) {
	var options easyKV.WatchOptions
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
		case <-stopChan:
			return options.WaitIndex, easyKV.ErrWatchCanceled
		case r := <-respChan:
			return r.waitIndex, r.err
		}
	}
}
