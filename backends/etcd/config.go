package etcd

import (
	"strings"

	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/backends/etcd/etcdv2"
	"github.com/HeavyHorst/remco/backends/etcd/etcdv3"
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
}

func (c *Config) NewClient() (backends.StoreClient, error) {
	log.Info("Backend nodes set to " + strings.Join(c.Nodes, ", "))
	if c.Version == 3 {
		return etcdv3.NewEtcdClient(c.Nodes, c.Cert, c.Key, c.CaCert, c.BasicAuth, c.Username, c.Password)
	} else {
		return etcdv2.NewEtcdClient(c.Nodes, c.Cert, c.Key, c.CaCert, c.BasicAuth, c.Username, c.Password)
	}
}
