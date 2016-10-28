/*
 * This file is part of easyKV.
 * Based on code from confd. https://github.com/kelseyhightower/confd
 * © 2013 Kelsey Hightower
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package zookeeper

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/HeavyHorst/easyKV"
	zk "github.com/tevino/go-zookeeper/zk"
)

// Client provides a wrapper around the zookeeper client
type Client struct {
	client *zk.Conn
}

// New returns an *zookeeper.Client with a connection to named machines.
// It returns an error if a connection to the cluster cannot be made.
func New(machines []string) (*Client, error) {
	c, _, err := zk.Connect(machines, time.Second)
	if err != nil {
		panic(err)
	}
	return &Client{c}, nil
}

// Close closes the client connection
func (c *Client) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

func nodeWalk(prefix string, c *Client, vars map[string]string) error {
	l, stat, err := c.client.Children(prefix)
	if err != nil {
		return err
	}

	if stat.NumChildren == 0 {
		b, _, err := c.client.Get(prefix)
		if err != nil {
			return err
		}
		vars[prefix] = string(b)

	} else {
		for _, key := range l {
			s := prefix + "/" + key
			_, stat, err := c.client.Exists(s)
			if err != nil {
				return err
			}
			if stat.NumChildren == 0 {
				b, _, err := c.client.Get(s)
				if err != nil {
					return err
				}
				vars[s] = string(b)
			} else {
				nodeWalk(s, c, vars)
			}
		}
	}
	return nil
}

// GetValues queries zookeeper for keys prefixed by prefix.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, v := range keys {
		v = strings.Replace(v, "/*", "", -1)
		_, _, err := c.client.Exists(v)
		if err != nil {
			return vars, err
		}
		if v == "/" {
			v = ""
		}
		err = nodeWalk(v, c, vars)
		if err != nil {
			return vars, err
		}
	}
	return vars, nil
}

type watchResponse struct {
	waitIndex uint64
	err       error
}

func (c *Client) watch(key string, respChan chan watchResponse, ctx context.Context) {
	_, _, keyWatcher, err := c.client.GetW(key)
	if err != nil {
		respChan <- watchResponse{0, err}
	}
	_, _, childWatcher, err := c.client.ChildrenW(key)
	if err != nil {
		respChan <- watchResponse{0, err}
	}

	for {
		select {
		case e := <-keyWatcher.EvtCh:
			if e.Type == zk.EventNodeDataChanged {
				respChan <- watchResponse{1, e.Err}
			}
		case e := <-childWatcher.EvtCh:
			if e.Type == zk.EventNodeChildrenChanged {
				respChan <- watchResponse{1, e.Err}
			}
		case <-ctx.Done():
			c.client.RemoveWatcher(childWatcher)
			c.client.RemoveWatcher(keyWatcher)
			return
		}
	}
}

// WatchPrefix watches a specific prefix for changes.
func (c *Client) WatchPrefix(prefix string, ctx context.Context, opts ...easyKV.WatchOption) (uint64, error) {
	var options easyKV.WatchOptions
	for _, o := range opts {
		o(&options)
	}

	// List the childrens first
	entries, err := c.GetValues([]string{prefix})
	if err != nil {
		return 0, err
	}

	respChan := make(chan watchResponse)
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(ctx)

	//watch all subfolders for changes
	watchMap := make(map[string]string)
	for k, _ := range entries {
		for _, v := range options.Keys {
			if strings.HasPrefix(k, v) {
				for dir := filepath.Dir(k); dir != "/"; dir = filepath.Dir(dir) {
					if _, ok := watchMap[dir]; !ok {
						watchMap[dir] = ""
						wg.Add(1)
						go func(dir string) {
							defer wg.Done()
							c.watch(dir, respChan, ctx)
						}(dir)
					}
				}
				break
			}
		}
	}

	//watch all keys in prefix for changes
	for k, _ := range entries {
		for _, v := range options.Keys {
			if strings.HasPrefix(k, v) {
				wg.Add(1)
				go func(k string) {
					defer wg.Done()
					c.watch(k, respChan, ctx)
				}(k)
				break
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return options.WaitIndex, nil
		case r := <-respChan:
			cancel()
			go func() {
				for range respChan {
				}
			}()
			wg.Wait()
			close(respChan)
			return r.waitIndex, r.err
		}
	}
}
