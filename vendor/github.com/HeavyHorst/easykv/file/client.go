/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package file

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"time"

	"github.com/HeavyHorst/easykv"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

// Client is a wrapper around the file client
type Client struct {
	filepath   string
	isURL      bool
	httpClient http.Client
}

// transport is a wrapper around any provided underlying transport that will
// also add any provided headers to any request made with this transport.
type transport struct {
	Headers             map[string]string
	UnderlyingTransport http.RoundTripper
}

// RoundTrip satisfies the RoundTripper interface for our custom transport above.
// If an underlying transport is not provided, we default to http.DefaultTransport.
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.Headers {
		req.Header.Add(k, v)
	}
	if t.UnderlyingTransport == nil {
		t.UnderlyingTransport = http.DefaultTransport
	}
	return t.UnderlyingTransport.RoundTrip(req)
}

// New returns a new FileClient
// The filepath can be a local path to a file or a remote http/https location.
func New(filepath string, opts ...Option) (*Client, error) {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	c := &Client{filepath: filepath}
	if strings.HasPrefix(filepath, "http://") || strings.HasPrefix(filepath, "https://") {
		c.isURL = true
		c.httpClient = http.Client{
			Timeout: 5 * time.Second,
			Transport: &transport{
				Headers: options.Headers,
			},
		}
	}
	return c, nil
}

// GetValues returns all key-value pairs from the yaml or json file where the
// keys begins with one of the prefixes specified in the keys array.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	yamlMap := make(map[interface{}]interface{})
	vars := make(map[string]string)
	kvs := make(map[string]string)

	var data []byte
	var err error
	if c.isURL {
		resp, err := c.httpClient.Get(c.filepath)
		if err != nil {
			return vars, err
		}
		defer resp.Body.Close()
		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return vars, err
		}
	} else {
		data, err = ioutil.ReadFile(c.filepath)
		if err != nil {
			return vars, err
		}
	}

	err = yaml.Unmarshal(data, &yamlMap)
	if err != nil {
		return vars, err
	}

	nodeWalk(yamlMap, "", vars)

	for _, k := range keys {
		for key, val := range vars {
			if strings.HasPrefix(key, k) {
				kvs[key] = val
			}
		}
	}

	return kvs, nil
}

// Close is only meant to fulfill the easykv.ReadWatcher interface.
// Does nothing.
func (c *Client) Close() {}

// nodeWalk recursively descends nodes, updating vars.
func nodeWalk(node interface{}, key string, vars map[string]string) error {
	switch node := node.(type) {
	case map[interface{}]interface{}:
		for k, v := range node {
			key := fmt.Sprintf("%s/%v", key, k)
			nodeWalk(v, key, vars)
		}
	case []interface{}:
		for i, j := range node {
			key := fmt.Sprintf("%s/%d", key, i)
			nodeWalk(j, key, vars)
		}
	default:
		vars[key] = fmt.Sprintf("%v", node)
	}
	return nil
}

// WatchPrefix watches the file for changes with fsnotify.
// Prefix, keys and waitIndex are only here to implement the StoreClient interface.
// WatchPrefix is only supported for local files. Remote files over http/https arent supported.
// Remote filesystems like nfs are also not supported.
func (c *Client) WatchPrefix(ctx context.Context, prefix string, opts ...easykv.WatchOption) (uint64, error) {
	if c.isURL {
		// watch is not supported for urls
		return 0, easykv.ErrWatchNotSupported
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return 0, err
	}
	defer watcher.Close()

	err = watcher.Add(c.filepath)
	if err != nil {
		return 0, err
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Remove == fsnotify.Remove {
				return 1, nil
			}
		case err := <-watcher.Errors:
			return 0, err
		case <-ctx.Done():
			return 0, easykv.ErrWatchCanceled
		}
	}
}
