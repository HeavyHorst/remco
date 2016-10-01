/*
 * This file is part of easyKV.
 * © 2016 The easyKV Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package kubernetesConfigMap

import (
	"strings"

	"github.com/HeavyHorst/easyKV"
	"k8s.io/client-go/1.4/kubernetes"
	"k8s.io/client-go/1.4/pkg/api"
	"k8s.io/client-go/1.4/pkg/api/v1"
	"k8s.io/client-go/1.4/pkg/watch"
	"k8s.io/client-go/1.4/rest"
	"k8s.io/client-go/1.4/tools/clientcmd"
)

// Client is a wrapper around the kubernetes client
type Client struct {
	client     *kubernetes.Clientset
	configName string
	namespace  string
}

func New(namespace, name string, opts ...Option) (easyKV.ReadWatcher, error) {
	var (
		options Options
		cfg     *rest.Config
		err     error
	)

	for _, o := range opts {
		o(&options)
	}

	if options.Config == "" {
		cfg, err = rest.InClusterConfig()
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags("", options.Config)
	}
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:     clientset,
		namespace:  namespace,
		configName: name,
	}, nil
}

// GetValues returns all key-value pairs from the config-map
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	cm, err := c.client.Core().ConfigMaps(c.namespace).Get(c.configName)
	if err != nil {
		return nil, err
	}
	vars := make(map[string]string)
	for _, key := range keys {
		k := strings.Replace(strings.TrimPrefix(key, "/"), "/", ".", -1)
		for cmKey, cmValue := range cm.Data {
			if strings.HasPrefix(cmKey, k) {
				vars["/"+strings.Replace(cmKey, ".", "/", -1)] = cmValue
			}
		}
	}

	return vars, nil
}

// WatchPrefix - watch a kubernetes namespace for changes
func (c *Client) WatchPrefix(prefix string, stopChan chan bool, opts ...easyKV.WatchOption) (uint64, error) {
	w, err := c.client.Core().ConfigMaps(c.namespace).Watch(api.ListOptions{})
	if err != nil {
		return 0, err
	}
	defer w.Stop()

	for {
		select {
		case <-stopChan:
			return 0, nil
		case event := <-w.ResultChan():
			if event.Type == watch.Modified {
				d := event.Object.(*v1.ConfigMap)
				if d.ObjectMeta.Name == c.configName {
					return 0, nil
				}
			}
		}
	}
}
