/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package backends

import (
	"github.com/HeavyHorst/easykv/etcd"
	berr "github.com/HeavyHorst/remco/pkg/backends/error"
	"github.com/HeavyHorst/remco/pkg/log"
	"github.com/HeavyHorst/remco/pkg/template"
)

// EtcdConfig represents the config for the etcd backend.
type EtcdConfig struct {
	// Nodes is list of backend nodes.
	Nodes []string

	// The backend URI scheme (http or https).
	// This is only used when the nodes are discovered via DNS srv records and the api level is 2.
	//
	// The default is http.
	Scheme string

	// The client cert file.
	ClientCert string `toml:"client_cert"`

	// The client key file.
	ClientKey string `toml:"client_key"`

	// The client CA key file.
	ClientCaKeys string `toml:"client_ca_keys"`

	// A DNS server record to discover the etcd nodes.
	SRVRecord SRVRecord `toml:"srv_record"`

	// The username for the basic_auth authentication.
	Username string

	// The password for the basic_auth authentication.
	Password string

	// The etcd api-level to use (2 or 3).
	//
	// The default is 2.
	Version int
	template.Backend

	// The timeout per request
	RequestTimeout int `toml:"request_timeout"`
}

// Connect creates a new etcd{2,3}Client and fills the underlying template.Backend with the etcd-Backend specific data.
func (c *EtcdConfig) Connect() (template.Backend, error) {
	if c == nil {
		return template.Backend{}, berr.ErrNilConfig
	}

	// use api version 2 if no version is specified
	if c.Version == 0 {
		c.Version = 2
	}

	if c.Version == 3 {
		c.Backend.Name = "etcdv3"
	} else {
		c.Backend.Name = "etcd"
	}

	// Use default request timeout of 3
	if c.RequestTimeout == 0 {
		c.RequestTimeout = 3
	}

	// No nodes are set but a SRVRecord is provided
	if len(c.Nodes) == 0 && c.SRVRecord != "" {
		var err error
		if c.Version != 2 {
			// no scheme required for etcdv3
			c.Scheme = ""
		} else if c.Scheme == "" {
			// etcd version is 2 and no scheme is provided
			// use http as default value
			c.Scheme = "http"
		}
		c.Nodes, err = c.SRVRecord.GetNodesFromSRV(c.Scheme)

		if err != nil {
			return c.Backend, err
		}
	}

	log.WithFields(
		"backend", c.Backend.Name,
		"nodes", c.Nodes,
	).Info("set backend nodes")

	client, err := etcd.New(c.Nodes,
		etcd.WithBasicAuth(etcd.BasicAuthOptions{
			Username: c.Username,
			Password: c.Password,
		}),
		etcd.WithTLSOptions(etcd.TLSOptions{
			ClientCert:   c.ClientCert,
			ClientKey:    c.ClientKey,
			ClientCaKeys: c.ClientCaKeys,
		}),
		etcd.WithVersion(c.Version),
		etcd.WithRequestTimeout(c.RequestTimeout))

	if err != nil {
		return c.Backend, err
	}

	c.Backend.ReadWatcher = client
	return c.Backend, nil
}

// return Backend config to allow modification before connect() for onetime param or similar
func (c *EtcdConfig) GetBackend() *template.Backend {
	return &c.Backend
}
