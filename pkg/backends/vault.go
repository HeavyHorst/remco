/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package backends

import (
	"github.com/HeavyHorst/easykv/vault"
	berr "github.com/HeavyHorst/remco/pkg/backends/error"
	"github.com/HeavyHorst/remco/pkg/log"
	"github.com/HeavyHorst/remco/pkg/template"
	"github.com/Sirupsen/logrus"
)

// VaultConfig represents the config for the vault backend.
type VaultConfig struct {
	Node         string
	AuthType     string `toml:"auth_type"`
	AppID        string `toml:"app_id"`
	UserID       string `toml:"user_id"`
	RoleID       string `toml:"role_id"`
	SecretID     string `toml:"secret_id"`
	Username     string
	Password     string
	AuthToken    string `toml:"auth_token"`
	ClientCert   string `toml:"client_cert"`
	ClientKey    string `toml:"client_key"`
	ClientCaKeys string `toml:"client_ca_keys"`
	template.Backend
}

// Connect creates a new vaultClient and fills the underlying template.Backend with the vault-Backend specific data.
func (c *VaultConfig) Connect() (template.Backend, error) {
	if c == nil {
		return template.Backend{}, berr.ErrNilConfig
	}

	c.Backend.Name = "vault"
	log.WithFields(logrus.Fields{
		"backend": c.Backend.Name,
		"nodes":   []string{c.Node},
	}).Info("Set backend nodes")

	tlsOps := vault.TLSOptions{
		ClientCert:   c.ClientCert,
		ClientKey:    c.ClientKey,
		ClientCaKeys: c.ClientCaKeys,
	}

	authOps := vault.BasicAuthOptions{
		Username: c.Username,
		Password: c.Password,
	}

	client, err := vault.New(c.Node, c.AuthType,
		vault.WithBasicAuth(authOps),
		vault.WithTLSOptions(tlsOps),
		vault.WithAppID(c.AppID),
		vault.WithUserID(c.UserID),
		vault.WithRoleID(c.RoleID),
		vault.WithSecretID(c.SecretID),
		vault.WithToken(c.AuthToken))

	if err != nil {
		return c.Backend, err
	}

	c.Backend.ReadWatcher = client

	if c.Backend.Watch {
		log.WithFields(logrus.Fields{
			"backend": c.Backend.Name,
		}).Warn("Watch is not supported, using interval instead")
		c.Backend.Watch = false
	}

	return c.Backend, nil
}
