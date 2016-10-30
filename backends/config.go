/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package backends

import (
	"github.com/HeavyHorst/remco/backends/consul"
	"github.com/HeavyHorst/remco/backends/env"
	"github.com/HeavyHorst/remco/backends/etcd"
	"github.com/HeavyHorst/remco/backends/file"
	"github.com/HeavyHorst/remco/backends/redis"
	"github.com/HeavyHorst/remco/backends/vault"
	"github.com/HeavyHorst/remco/backends/zookeeper"
	"github.com/HeavyHorst/remco/template"
)

type Config struct {
	Etcd      *etcd.Config
	File      *file.Config
	Env       *env.Config
	Consul    *consul.Config
	Vault     *vault.Config
	Redis     *redis.Config
	Zookeeper *zookeeper.Config
}

func (c *Config) GetBackends() []template.BackendConfig {
	return []template.BackendConfig{
		c.Etcd,
		c.File,
		c.Env,
		c.Consul,
		c.Vault,
		c.Redis,
		c.Zookeeper,
	}
}
