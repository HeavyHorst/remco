/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package backends

import (
	"github.com/HeavyHorst/easykv/zookeeper"
	berr "github.com/HeavyHorst/remco/pkg/backends/error"
	"github.com/HeavyHorst/remco/pkg/log"
	"github.com/HeavyHorst/remco/pkg/template"
	"github.com/sirupsen/logrus"
)

// ZookeeperConfig represents the config for the consul backend.
type ZookeeperConfig struct {
	// A list of zookeeper nodes.
	Nodes []string

	// A DNS server record to discover the zookeeper nodes.
	SRVRecord SRVRecord `toml:"srv_record"`
	template.Backend
}

// Connect creates a new zookeeperClient and fills the underlying template.Backend with the zookeeper-Backend specific data.
func (c *ZookeeperConfig) Connect() (template.Backend, error) {
	if c == nil {
		return template.Backend{}, berr.ErrNilConfig
	}

	c.Backend.Name = "zookeeper"

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
	}).Info("set backend nodes")

	client, err := zookeeper.New(c.Nodes)
	if err != nil {
		return c.Backend, err
	}

	c.Backend.ReadWatcher = client
	return c.Backend, nil
}
