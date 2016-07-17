package vault

import (
	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/confd/backends/vault"
)

type Config struct {
	Node         string
	AuthType     string `toml:"auth_type"`
	AppID        string `toml:"app_id"`
	UserID       string `toml:"user_id"`
	Username     string
	Password     string
	AuthToken    string `toml:"auth_token"`
	ClientCert   string `toml:"client_cert"`
	ClientKey    string `toml:"client_key"`
	ClientCaKeys string `toml:"client_ca_keys"`
	template.Backend
}

func (c *Config) Connect() (backends.Store, error) {
	log.WithFields(logrus.Fields{
		"backend": "vault",
		"nodes":   []string{c.Node},
	}).Info("Set backend nodes")

	vaultConfig := map[string]string{
		"app-id":   c.AppID,
		"user-id":  c.UserID,
		"username": c.Username,
		"password": c.Password,
		"token":    c.AuthToken,
		"cert":     c.ClientCert,
		"key":      c.ClientKey,
		"caCert":   c.ClientCaKeys,
	}
	client, err := vault.New(c.Node, c.AuthType, vaultConfig)
	if err != nil {
		return backends.Store{}, err
	}
	c.Backend.StoreClient = client
	c.Backend.Name = "vault"
	return backends.Store{
		Name:   c.Backend.Name,
		Client: client,
	}, nil
}
