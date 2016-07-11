package consul

import (
	"strings"

	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/kelseyhightower/confd/backends/consul"
)

type Config struct {
	Nodes  []string
	Scheme string
	Cert   string
	Key    string
	CaCert string
	template.StoreConfig
}

func (c *Config) Connect() (backends.Store, error) {
	log.Info("consul backend nodes set to " + strings.Join(c.Nodes, ", "))
	client, err := consul.New(c.Nodes, c.Scheme, c.Cert, c.Key, c.CaCert)
	if err != nil {
		return backends.Store{}, err
	}
	c.StoreConfig.StoreClient = client
	c.StoreConfig.Name = "consul"
	return backends.Store{
		Name:   c.StoreConfig.Name,
		Client: client,
	}, nil
}
