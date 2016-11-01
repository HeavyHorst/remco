/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package mock

import (
	"context"

	"github.com/HeavyHorst/easyKV"
)

// Client is the mock client
type Client struct {
	Err error
}

// New creates a new mock client, err will be returned by all methods
func New(err error) (*Client, error) {
	return &Client{Err: err}, nil
}

// GetValues mock
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	return make(map[string]string), c.Err
}

// Close mock
func (c *Client) Close() {
	return
}

// WatchPrefix mock
func (c *Client) WatchPrefix(prefix string, ctx context.Context, opts ...easyKV.WatchOption) (uint64, error) {
	return 0, c.Err
}
