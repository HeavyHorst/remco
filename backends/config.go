/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package backends

import (
	"github.com/HeavyHorst/remco/backends/plugin"

	"github.com/HeavyHorst/remco/template"
)

type Config struct {
	Etcd      *EtcdConfig
	File      *FileConfig
	Env       *EnvConfig
	Consul    *ConsulConfig
	Vault     *VaultConfig
	Redis     *RedisConfig
	Zookeeper *ZookeeperConfig
	Mock      *MockConfig
	Plugin    []plugin.Plugin
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
		c.Mock,
	}
}
