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
	"time"

	"github.com/HeavyHorst/easyKV"
)

// Client is the mock client
type Client struct {
	Err  error
	Data map[string]string
}

// New creates a new mock client, err will be returned by all methods
func New(err error, data map[string]string) (*Client, error) {
	return &Client{
		Err:  err,
		Data: data,
	}, nil
}

// GetValues mock
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	return c.Data, c.Err
}

// Close mock
func (c *Client) Close() {
	return
}

// WatchPrefix mock
func (c *Client) WatchPrefix(ctx context.Context, prefix string, opts ...easyKV.WatchOption) (uint64, error) {
	time.Sleep(2 * time.Second)
	return 0, c.Err
}
