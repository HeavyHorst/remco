/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package file

// Options contains all (possibly optional) values that are needed to fetch
// JSON or YAML files (either locally or remotely).
type Options struct {
	Headers map[string]string
}

// Option configures the file client.
type Option func(*Options)

// WithHeaders sets the headers for the HTTP request made when fetching
// remote files.
func WithHeaders(headers map[string]string) Option {
	return func(o *Options) {
		o.Headers = headers
	}
}
