package file

import (
	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/Sirupsen/logrus"
)

type Config struct {
	Filepath string
	template.Backend
}

func (c *Config) Connect() (template.Backend, error) {
	if c == nil {
		return template.Backend{}, backends.ErrNilConfig
	}

	log.WithFields(logrus.Fields{
		"backend":  "file",
		"filepath": c.Filepath,
	}).Info("Set filepath")
	client, err := NewFileClient(c.Filepath)
	if err != nil {
		return c.Backend, err
	}
	c.Backend.StoreClient = client
	c.Backend.Name = "file"
	return c.Backend, nil
}
