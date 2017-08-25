/*
 * This file is part of easyKV.
 * Based on code from confd.
 * https://github.com/kelseyhightower/confd/blob/2cacfab234a5d61be4cd88b9e97bee44437c318d/backends/env/client.go
 * © 2013 Kelsey Hightower
 *
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package env

import (
	"context"
	"os"
	"strings"

	"github.com/HeavyHorst/easykv"
)

var replacer = strings.NewReplacer("/", "_")
var cleanReplacer = strings.NewReplacer("_", "/")

// Client provides a shell for the env client
type Client struct{}

// New returns a new client
func New() (*Client, error) {
	return &Client{}, nil
}

// Close is only meant to fulfill the easykv.ReadWatcher interface.
// Does nothing.
func (c *Client) Close() {}

// GetValues is used to lookup all keys with a prefix.
// Several prefixes can be specified in the keys array.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	allEnvVars := os.Environ()
	envMap := make(map[string]string)
	for _, e := range allEnvVars {
		index := strings.Index(e, "=")
		envMap[e[:index]] = e[index+1:]
	}
	vars := make(map[string]string)
	for _, key := range keys {
		k := transform(key)
		for envKey, envValue := range envMap {
			if strings.HasPrefix(envKey, k) {
				vars[clean(envKey)] = envValue
			}
		}
	}

	return vars, nil
}

func transform(key string) string {
	k := strings.TrimPrefix(key, "/")
	return strings.ToUpper(replacer.Replace(k))
}

func clean(key string) string {
	newKey := "/" + key
	return cleanReplacer.Replace(strings.ToLower(newKey))
}

// WatchPrefix - not implemented at the moment
func (c *Client) WatchPrefix(ctx context.Context, prefix string, opts ...easykv.WatchOption) (uint64, error) {
	return 0, easykv.ErrWatchNotSupported
}
