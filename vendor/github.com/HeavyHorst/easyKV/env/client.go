/*
 * This file is part of easyKV.
 * Based on code from confd. https://github.com/kelseyhightower/confd
 * © 2013 Kelsey Hightower
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

	"github.com/HeavyHorst/easyKV"
)

var replacer = strings.NewReplacer("/", "_")
var cleanReplacer = strings.NewReplacer("_", "/")

// Client provides a shell for the env client
type Client struct{}

// New returns a new client
func New() (*Client, error) {
	return &Client{}, nil
}

// Close closes the client connection
func (c *Client) Close() {
	return
}

// GetValues queries the environment for keys
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
func (c *Client) WatchPrefix(prefix string, ctx context.Context, opts ...easyKV.WatchOption) (uint64, error) {
	return 0, easyKV.ErrWatchNotSupported
}
