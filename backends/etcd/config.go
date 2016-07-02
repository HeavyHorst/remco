package etcd

import (
	"strings"

	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/backends/etcd/etcdv2"
	"github.com/HeavyHorst/remco/backends/etcd/etcdv3"
	"github.com/HeavyHorst/remco/template"
	"github.com/cloudflare/cfssl/log"
)

type Config struct {
	Nodes     []string
	Cert      string
	Key       string
	CaCert    string
	BasicAuth bool
	Username  string
	Password  string
	Version   int
	template.StoreConfig
}

func (c *Config) Connect() (backends.Store, error) {
	log.Info("Backend nodes set to " + strings.Join(c.Nodes, ", "))
	var client backends.StoreClient
	var err error

	if c.Version == 3 {
		client, err = etcdv3.NewEtcdClient(c.Nodes, c.Cert, c.Key, c.CaCert, c.BasicAuth, c.Username, c.Password)
		c.StoreConfig.Name = "etcdv3"
	} else {
		client, err = etcdv2.NewEtcdClient(c.Nodes, c.Cert, c.Key, c.CaCert, c.BasicAuth, c.Username, c.Password)
		c.StoreConfig.Name = "etcd"
	}

	if err != nil {
		return backends.Store{}, err
	}

	c.StoreConfig.StoreClient = client
	return backends.Store{
		Name:   c.StoreConfig.Name,
		Client: client,
	}, nil
}
