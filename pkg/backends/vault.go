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
	"github.com/sirupsen/logrus"
)

// VaultConfig represents the config for the vault backend.
type VaultConfig struct {
	// The address of the vault server.
	Node string

	// The vault authentication type.
	//   (token, approle, app-id, userpass, github, cert, kubernetes)
	AuthType string `toml:"auth_type"`

	// The vault app ID.
	// Only used with auth_type=app-id.
	AppID string `toml:"app_id"`
	// The vault user ID.
	// Only used with auth_type=app-id.
	UserID string `toml:"user_id"`

	// The vault RoleID.
	// Only used with auth_type=approle and kubernetes.
	RoleID string `toml:"role_id"`
	// The vault SecretID.
	// Only used with auth_type=approle.
	SecretID string `toml:"secret_id"`

	// The username for the userpass authentication.
	Username string
	// The password for the userpass authentication.
	Password string

	// The vault authentication token. Only used with auth_type=token and github.
	AuthToken string `toml:"auth_token"`

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
	}).Info("set backend nodes")

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

// return Backend config to allow modification before connect() for onetime param or similar
func (c *VaultConfig) GetBackend() *template.Backend {
	return &c.Backend
}
