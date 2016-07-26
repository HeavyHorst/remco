package vault

import (
	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/confd/backends/vault"
)

// Config represents the config for the vault backend.
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

// Connect creates a new vaultClient and fills the underlying template.Backend with the vault-Backend specific data.
func (c *Config) Connect() (template.Backend, error) {
	if c == nil {
		return template.Backend{}, backends.ErrNilConfig
	}

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
		return c.Backend, err
	}
	c.Backend.StoreClient = client
	c.Backend.Name = "vault"

	if c.Backend.Watch {
		log.WithFields(logrus.Fields{
			"backend": "vault",
		}).Warn("Watch is not supported, using interval instead")
		c.Backend.Watch = false
	}

	return c.Backend, nil
}
