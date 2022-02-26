/*
 * This file is part of remco.
 * © 2022 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package backends

import (
	"github.com/HeavyHorst/easykv/nats"
	berr "github.com/HeavyHorst/remco/pkg/backends/error"
	"github.com/HeavyHorst/remco/pkg/log"
	"github.com/HeavyHorst/remco/pkg/template"
)

// NatsConfig represents the config for the nats backend.
type NatsConfig struct {
	// Nodes is list of backend nodes.
	Nodes []string

	// The Nats kv bucket
	Bucket string

	// The username for the basic_auth authentication.
	Username string

	// The password for the basic_auth authentication.
	Password string

	// The token for the token auth method
	Token string

	// The path to an NATS 2.0 and NATS NGS compatible user credentials file
	Creds string

	template.Backend
}

// Connect creates a new etcd{2,3}Client and fills the underlying template.Backend with the etcd-Backend specific data.
func (c *NatsConfig) Connect() (template.Backend, error) {
	if c == nil {
		return template.Backend{}, berr.ErrNilConfig
	}

	c.Backend.Name = "nats"

	log.WithFields(
		"backend", c.Backend.Name,
		"nodes", c.Nodes,
	).Info("set backend nodes")

	client, err := nats.New(c.Nodes, c.Bucket, nats.WithBasicAuth(nats.BasicAuthOptions{
		Username: c.Username,
		Password: c.Password,
	}), nats.WithToken(c.Token), nats.WithCredentials(c.Creds))

	if err != nil {
		return c.Backend, err
	}

	c.Backend.ReadWatcher = client
	return c.Backend, nil
}

// return Backend config to allow modification before connect() for onetime param or similar
func (c *NatsConfig) GetBackend() *template.Backend {
	return &c.Backend
}
