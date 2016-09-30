/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package configmap

import (
	"github.com/HeavyHorst/easyKV/kubernetesConfigMap"
	"github.com/HeavyHorst/remco/backends"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/Sirupsen/logrus"
)

// Config represents the config for the configmap backend.
type Config struct {
	KubeConfig string
	Name       string
	Namespace  string
	template.Backend
}

// Connect creates a new kubernetesClient and fills the underlying template.Backend with the configmap-Backend specific data.
func (c *Config) Connect() (template.Backend, error) {
	if c == nil {
		return template.Backend{}, backends.ErrNilConfig
	}

	log.WithFields(logrus.Fields{
		"backend": "configMap",
	}).Info("Set backend")

	client, err := kubernetesConfigMap.New(c.Namespace, c.Name, kubernetesConfigMap.WithConfig(c.KubeConfig))

	if err != nil {
		return c.Backend, err
	}

	c.Backend.ReadWatcher = client
	c.Backend.Name = "configMap"

	if c.Backend.Watch {
		log.WithFields(logrus.Fields{
			"backend": "configMap",
		}).Warn("Watch is not supported, using interval instead")
		c.Backend.Watch = false
	}

	return c.Backend, nil
}
