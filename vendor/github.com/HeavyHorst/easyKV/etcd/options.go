/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package etcd

// Options contains all values that are needed to connect to etcd
type Options struct {
	Nodes   []string
	Version int
	TLS     TLSOptions
	Auth    BasicAuthOptions
}

// TLSOptions contains all certificates and keys
type TLSOptions struct {
	ClientCert   string
	ClientKey    string
	ClientCaKeys string
}

// BasicAuthOptions contains options regarding to basic authentication
type BasicAuthOptions struct {
	BasicAuth bool
	Username  string
	Password  string
}

// Option configures the etcd client
type Option func(*Options)

// WithTLSOptions sets the TLSOptions
func WithTLSOptions(tls TLSOptions) Option {
	return func(o *Options) {
		o.TLS = tls
	}
}

// WithBasicAuth enables the basic authentication and sets the username and password
func WithBasicAuth(b BasicAuthOptions) Option {
	return func(o *Options) {
		o.Auth = b
	}
}

// WithVersion sets the etcd api level. Valid levels are 2 and 3.
func WithVersion(v int) Option {
	return func(o *Options) {
		o.Version = v
	}
}
