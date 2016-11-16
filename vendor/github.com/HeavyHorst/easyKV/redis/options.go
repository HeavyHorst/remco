/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package redis

// Option configures the redis client
type Option func(*Client)

// WithPassword sets the redis password
func WithPassword(pw string) Option {
	return func(o *Client) {
		o.password = pw
	}
}

// WithDatabase configures the redis database
func WithDatabase(db int) Option {
	return func(o *Client) {
		o.db = db
	}
}
