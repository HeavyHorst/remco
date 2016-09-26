/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package file

import (
	"io/ioutil"

	"github.com/HeavyHorst/easyKV"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

// Client is a wrapper around the file client
type Client struct {
	filepath string
}

// New returns a new FileClient
func New(filepath string) (*Client, error) {
	return &Client{filepath}, nil
}

// GetValues returns all key-value pairs from the yaml or json file.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	yamlMap := make(map[interface{}]interface{})
	vars := make(map[string]string)

	data, err := ioutil.ReadFile(c.filepath)
	if err != nil {
		return vars, err
	}
	err = yaml.Unmarshal(data, &yamlMap)
	if err != nil {
		return vars, err
	}

	nodeWalk(yamlMap, "", vars)

	return vars, nil
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
func (c *Client) WatchPrefix(prefix string, stopChan chan bool, opts ...easyKV.WatchOption) (uint64, error) {
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
		case <-stopChan:
			return 0, nil
		}
	}
}
