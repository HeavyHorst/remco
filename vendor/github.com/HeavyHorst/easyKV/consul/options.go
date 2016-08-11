/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package consul

// Options contains all values that are needed to connect to consul
type Options struct {
	Scheme string
	TLS    TLSOptions
}

// TLSOptions contains all certificates and keys
type TLSOptions struct {
	ClientCert   string
	ClientKey    string
	ClientCaKeys string
}

// Option configures the consul client
type Option func(*Options)

// WithScheme sets the consul uri scheme
func WithScheme(scheme string) Option {
	return func(o *Options) {
		o.Scheme = scheme
	}
}

// WithTLSOptions sets the TLSOptions
func WithTLSOptions(tls TLSOptions) Option {
	return func(o *Options) {
		o.TLS = tls
	}
}
