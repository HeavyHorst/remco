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
	"io/ioutil"
	"net/http"
	"strings"

	"time"

	"github.com/HeavyHorst/easyKV"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

// Client is a wrapper around the file client
type Client struct {
	filepath   string
	isURL      bool
	httpClient http.Client
}

// New returns a new FileClient
func New(filepath string) (*Client, error) {
	c := &Client{filepath: filepath}
	if strings.HasPrefix(filepath, "http://") || strings.HasPrefix(filepath, "https://") {
		c.isURL = true
		c.httpClient = http.Client{
			Timeout: 5 * time.Second,
		}
	}
	return c, nil
}

// GetValues returns all key-value pairs from the yaml or json file.
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

// Close closes the client connection
func (c *Client) Close() {
	return
}

// nodeWalk recursively descends nodes, updating vars.
func nodeWalk(node map[interface{}]interface{}, key string, vars map[string]string) error {
	for k, v := range node {
		key := key + "/" + k.(string)

		switch v.(type) {
		case map[interface{}]interface{}:
			nodeWalk(v.(map[interface{}]interface{}), key, vars)
		case []interface{}:
			for _, j := range v.([]interface{}) {
				switch j.(type) {
				case map[interface{}]interface{}:
					nodeWalk(j.(map[interface{}]interface{}), key, vars)
				case string:
					vars[key+"/"+j.(string)] = ""
				}
			}
		case string:
			vars[key] = v.(string)
		}
	}
	return nil
}

// WatchPrefix watches the file for changes with fsnotify.
// Prefix, keys and waitIndex are only here to implement the StoreClient interface.
func (c *Client) WatchPrefix(ctx context.Context, prefix string, opts ...easyKV.WatchOption) (uint64, error) {
	if c.isURL {
		// watch is not supported for urls
		return 0, easyKV.ErrWatchNotSupported
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
			return 0, easyKV.ErrWatchCanceled
		}
	}
}
