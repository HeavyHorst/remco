/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package backends

import (
	"github.com/HeavyHorst/easykv/redis"
	berr "github.com/HeavyHorst/remco/pkg/backends/error"
	"github.com/HeavyHorst/remco/pkg/log"
	"github.com/HeavyHorst/remco/pkg/template"
)

// RedisConfig represents the config for the redis backend.
type RedisConfig struct {
	// A list of backend nodes.
	Nodes []string

	// A DNS server record to discover the redis nodes.
	SRVRecord SRVRecord `toml:"srv_record"`

	// The redis password.
	Password string

	// The redis database.
	Database int

	template.Backend
}

// Connect creates a new redisClient and fills the underlying template.Backend with the redis-Backend specific data.
func (c *RedisConfig) Connect() (template.Backend, error) {
	if c == nil {
		return template.Backend{}, berr.ErrNilConfig
	}

	c.Backend.Name = "redis"

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

	client, err := redis.New(c.Nodes, redis.WithPassword(c.Password), redis.WithDatabase(c.Database))
	if err != nil {
		return c.Backend, err
	}

	c.Backend.ReadWatcher = client

	if c.Backend.Watch {
		log.WithFields(
			"backend", c.Backend.Name,
		).Warn("Watch is not supported, using interval instead")
		c.Backend.Watch = false
	}

	return c.Backend, nil
}

// return Backend config to allow modification before connect() for onetime param or similar
func (c *RedisConfig) GetBackend() *template.Backend {
	return &c.Backend
}
