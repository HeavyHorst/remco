/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package backends

import (
	"github.com/HeavyHorst/easyKV/etcd"
	berr "github.com/HeavyHorst/remco/backends/error"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/Sirupsen/logrus"
)

// EtcdConfig represents the config for the etcd backend.
type EtcdConfig struct {
	Nodes        []string
	Scheme       string    //only needed for etcdv2 with SRVRecord
	ClientCert   string    `toml:"client_cert"`
	ClientKey    string    `toml:"client_key"`
	ClientCaKeys string    `toml:"client_ca_keys"`
	SRVRecord    SRVRecord `toml:"srv_record"`
	Username     string
	Password     string
	Version      int
	template.Backend
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

	log.WithFields(logrus.Fields{
		"backend": c.Backend.Name,
		"nodes":   c.Nodes,
	}).Info("Set backend nodes")

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
		etcd.WithVersion(c.Version))

	if err != nil {
		return c.Backend, err
	}

	c.Backend.ReadWatcher = client
	return c.Backend, nil
}
