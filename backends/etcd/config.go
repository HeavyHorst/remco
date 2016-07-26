package etcd

import (
	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/backends/etcd/etcdv2"
	"github.com/HeavyHorst/remco/backends/etcd/etcdv3"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/Sirupsen/logrus"
)

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

func (c *Config) Connect() (template.Backend, error) {
	if c == nil {
		return template.Backend{}, backends.ErrNilConfig
	}

	//log.Info("etcd backend nodes set to " + strings.Join(c.Nodes, ", "))
	log.WithFields(logrus.Fields{
		"backend": "etcd",
		"nodes":   c.Nodes,
	}).Info("Set backend nodes")
	var client backends.StoreClient
	var err error

	if c.Version == 3 {
		client, err = etcdv3.NewEtcdClient(c.Nodes, c.ClientCert, c.ClientKey, c.ClientCaKeys, c.BasicAuth, c.Username, c.Password)
		c.Backend.Name = "etcdv3"
	} else {
		client, err = etcdv2.NewEtcdClient(c.Nodes, c.ClientCert, c.ClientKey, c.ClientCaKeys, c.BasicAuth, c.Username, c.Password)
		c.Backend.Name = "etcd"
	}

	if err != nil {
		return c.Backend, err
	}

	c.Backend.StoreClient = client
	return c.Backend, nil
}
