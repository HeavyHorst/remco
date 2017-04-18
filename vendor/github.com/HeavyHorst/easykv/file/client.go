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

// New returns a new FileClient
// The filepath can be a local path to a file or a remote http/https location.
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
func (c *Client) Close() {
	return
}

// nodeWalk recursively descends nodes, updating vars.
func nodeWalk(node map[interface{}]interface{}, key string, vars map[string]string) error {
	for k, v := range node {
		key := key + "/" + k.(string)

		switch v := v.(type) {
		case map[interface{}]interface{}:
			nodeWalk(v, key, vars)
		case []interface{}:
			for _, j := range v {
				switch j := j.(type) {
				case map[interface{}]interface{}:
					nodeWalk(j, key, vars)
				case string:
					vars[key+"/"+j] = ""
				}
			}
		case string:
			vars[key] = v
		}
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
