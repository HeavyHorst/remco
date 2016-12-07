/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package redis

import (
	"github.com/HeavyHorst/easyKV/redis"
	berr "github.com/HeavyHorst/remco/backends/error"
	"github.com/HeavyHorst/remco/backends/srvRecord"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/Sirupsen/logrus"
)

// Config represents the config for the redis backend.
type Config struct {
	Nodes     []string
	SRVRecord srvRecord.Record `toml:"srv_record"`
	Password  string
	Database  int
	template.Backend
}

// Connect creates a new redisClient and fills the underlying template.Backend with the redis-Backend specific data.
func (c *Config) Connect() (template.Backend, error) {
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

	log.WithFields(logrus.Fields{
		"backend": c.Backend.Name,
		"nodes":   c.Nodes,
	}).Info("Set backend nodes")

	client, err := redis.New(c.Nodes, redis.WithPassword(c.Password), redis.WithDatabase(c.Database))
	if err != nil {
		return c.Backend, err
	}

	c.Backend.ReadWatcher = client

	if c.Backend.Watch {
		log.WithFields(logrus.Fields{
			"backend": c.Backend.Name,
		}).Warn("Watch is not supported, using interval instead")
		c.Backend.Watch = false
	}

	return c.Backend, nil
}
