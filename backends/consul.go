/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package backends

import (
	"github.com/HeavyHorst/easyKV/consul"
	berr "github.com/HeavyHorst/remco/backends/error"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/Sirupsen/logrus"
)

// ConsulConfig represents the config for the consul backend.
type ConsulConfig struct {
	Nodes        []string
	Scheme       string
	SRVRecord    SRVRecord `toml:"srv_record"`
	ClientCert   string    `toml:"client_cert"`
	ClientKey    string    `toml:"client_key"`
	ClientCaKeys string    `toml:"client_ca_keys"`
	template.Backend
}

// Connect creates a new consulClient and fills the underlying template.Backend with the consul-Backend specific data.
func (c *ConsulConfig) Connect() (template.Backend, error) {
	if c == nil {
		return template.Backend{}, berr.ErrNilConfig
	}
	c.Backend.Name = "consul"

	// No nodes are set but a SRVRecord is provided
	if len(c.Nodes) == 0 && c.SRVRecord != "" {
		var err error
		c.Nodes, err = c.SRVRecord.GetNodesFromSRV("")
		if err != nil {
			return c.Backend, err
		}
	}

	log.WithFields(logrus.Fields{
		"backend": c.Backend.Name,
		"nodes":   c.Nodes,
	}).Info("Set backend nodes")

	client, err := consul.New(c.Nodes, consul.WithScheme(c.Scheme), consul.WithTLSOptions(consul.TLSOptions{
		ClientCert:   c.ClientCert,
		ClientKey:    c.ClientKey,
		ClientCaKeys: c.ClientCaKeys,
	}))

	if err != nil {
		return c.Backend, err
	}

	c.Backend.ReadWatcher = client

	return c.Backend, nil
}
