/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package consul

import (
	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/confd/backends/consul"
)

// Config represents the config for the consul backend.
type Config struct {
	Nodes        []string
	Scheme       string
	ClientCert   string `toml:"client_cert"`
	ClientKey    string `toml:"client_key"`
	ClientCaKeys string `toml:"client_ca_keys"`
	template.Backend
}

// Connect creates a new consulClient and fills the underlying template.Backend with the consul-Backend specific data.
func (c *Config) Connect() (template.Backend, error) {
	if c == nil {
		return template.Backend{}, backends.ErrNilConfig
	}

	log.WithFields(logrus.Fields{
		"backend": "consul",
		"nodes":   c.Nodes,
	}).Info("Set backend nodes")

	client, err := consul.New(c.Nodes, c.Scheme, c.ClientCert, c.ClientKey, c.ClientCaKeys)
	if err != nil {
		return c.Backend, err
	}

	c.Backend.StoreClient = client
	c.Backend.Name = "consul"

	return c.Backend, nil
}
