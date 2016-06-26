package consul

import (
	"strings"

	"github.com/HeavyHorst/remco/backends"
	"github.com/cloudflare/cfssl/log"
	"github.com/kelseyhightower/confd/backends/consul"
)

type Config struct {
	Nodes  []string
	Scheme string
	Cert   string
	Key    string
	CaCert string
}

func (c *Config) NewClient() (backends.StoreClient, error) {
	log.Info("Backend nodes set to " + strings.Join(c.Nodes, ", "))
	return consul.New(c.Nodes, c.Scheme, c.Cert, c.Key, c.CaCert)
}
