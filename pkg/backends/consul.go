/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package backends

import (
	"github.com/HeavyHorst/easykv/consul"
	berr "github.com/HeavyHorst/remco/pkg/backends/error"
	"github.com/HeavyHorst/remco/pkg/log"
	"github.com/HeavyHorst/remco/pkg/template"
)

// ConsulConfig represents the config for the consul backend.
type ConsulConfig struct {
	// Nodes is a list of backend nodes.
	// A connection will only be established to the first server in the list.
	Nodes []string

	// The backend URI scheme (http or https).
	Scheme string

	// A DNS server record to discover the consul nodes.
	SRVRecord SRVRecord `toml:"srv_record"`

	// The client cert file.
	ClientCert string `toml:"client_cert"`

	// The client key file.
	ClientKey string `toml:"client_key"`

	//The client CA key file.
	ClientCaKeys string `toml:"client_ca_keys"`

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

	log.WithFields(
		"backend", c.Backend.Name,
		"nodes", c.Nodes,
	).Info("set backend nodes")

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

// return Backend config to allow modification before connect() for onetime param or similar
func (c *ConsulConfig) GetBackend() *template.Backend {
	return &c.Backend
}
