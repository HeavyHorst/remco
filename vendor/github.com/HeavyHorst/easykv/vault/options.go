/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package vault

// Options contains all values that are needed to connect to vault.
type Options struct {
	RoleID   string
	SecretID string
	AppID    string
	UserID   string
	Token    string
	TLS      TLSOptions
	Auth     BasicAuthOptions
}

// BasicAuthOptions contains options regarding to basic authentication.
type BasicAuthOptions struct {
	Username string
	Password string
}

// TLSOptions contains all certificates and keys.
type TLSOptions struct {
	ClientCert   string
	ClientKey    string
	ClientCaKeys string
}

// Option configures the vault client.
type Option func(*Options)

// WithAppID sets the AppID (app-id auth method).
func WithAppID(id string) Option {
	return func(o *Options) {
		o.AppID = id
	}
}

// WithUserID sets the UserID (app-id auth method).
func WithUserID(id string) Option {
	return func(o *Options) {
		o.UserID = id
	}
}

// WithRoleID sets the RoleID (approle auth method).
func WithRoleID(id string) Option {
	return func(o *Options) {
		o.RoleID = id
	}
}

// WithSecretID sets the ScretID (approle auth method).
func WithSecretID(id string) Option {
	return func(o *Options) {
		o.SecretID = id
	}
}

// WithToken sets the token (toke auth method).
func WithToken(token string) Option {
	return func(o *Options) {
		o.Token = token
	}
}

// WithTLSOptions sets the TLSOptions.
func WithTLSOptions(tls TLSOptions) Option {
	return func(o *Options) {
		o.TLS = tls
	}
}

// WithBasicAuth enables the basic authentication and sets the username and password.
func WithBasicAuth(b BasicAuthOptions) Option {
	return func(o *Options) {
		o.Auth = b
	}
}
