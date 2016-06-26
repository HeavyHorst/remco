package file

import (
	"github.com/HeavyHorst/remco/backends"
	"github.com/cloudflare/cfssl/log"
)

type Config struct {
	Filepath string
}

func (c *Config) NewClient() (backends.StoreClient, error) {
	log.Info("Filepath set to " + c.Filepath)
	return NewFileClient(c.Filepath)
}
