/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package etcd

import (
	"github.com/HeavyHorst/easyKV"
	"github.com/HeavyHorst/easyKV/etcd"
	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/Sirupsen/logrus"
)

// Config represents the config for the etcd backend.
type Config struct {
	Nodes        []string
	ClientCert   string `toml:"client_cert"`
	ClientKey    string `toml:"client_key"`
	ClientCaKeys string `toml:"client_ca_keys"`
	BasicAuth    bool   `toml:"basic_auth"`
	Username     string
	Password     string
	Version      int
	template.Backend
}

// Connect creates a new etcd{2,3}Client and fills the underlying template.Backend with the etcd-Backend specific data.
func (c *Config) Connect() (template.Backend, error) {
	if c == nil {
		return template.Backend{}, backends.ErrNilConfig
	}

	log.WithFields(logrus.Fields{
		"backend": "etcd",
		"nodes":   c.Nodes,
	}).Info("Set backend nodes")
	var client easyKV.ReadWatcher
	var err error

	client, err = etcd.NewEtcdClient(c.Nodes,
		etcd.WithBasicAuth(etcd.BasicAuthOptions{
			Username:  c.Username,
			Password:  c.Password,
			BasicAuth: c.BasicAuth,
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

	if c.Version == 3 {
		c.Backend.Name = "etcdv3"
	} else {
		c.Backend.Name = "etcd"
	}

	c.Backend.ReadWatcher = client
	return c.Backend, nil
}
