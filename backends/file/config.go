package file

import (
	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/template"
	"github.com/cloudflare/cfssl/log"
)

type Config struct {
	Filepath string
	template.StoreConfig
}

func (c *Config) Connect() (backends.Store, error) {
	log.Info("Filepath set to " + c.Filepath)
	client, err := NewFileClient(c.Filepath)
	if err != nil {
		return backends.Store{}, err
	}
	c.StoreConfig.StoreClient = client
	c.StoreConfig.Name = "file"
	return backends.Store{
		Name:   "file",
		Client: client,
	}, nil
}
