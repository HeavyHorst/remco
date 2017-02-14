/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/HeavyHorst/remco/pkg/backends"
	"github.com/HeavyHorst/remco/pkg/backends/plugin"
	"github.com/HeavyHorst/remco/pkg/log"
	"github.com/HeavyHorst/remco/pkg/template"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

// BackendConfigs holds every individually backend config.
// The values are filled with data from the configuration file.
type BackendConfigs struct {
	Etcd      *backends.EtcdConfig
	File      *backends.FileConfig
	Env       *backends.EnvConfig
	Consul    *backends.ConsulConfig
	Vault     *backends.VaultConfig
	Redis     *backends.RedisConfig
	Zookeeper *backends.ZookeeperConfig
	Mock      *backends.MockConfig
	Plugin    []plugin.Plugin
}

// GetBackends returns a slice with all BackendConfigs for easy iteration.
func (c *BackendConfigs) GetBackends() template.BackendConnectors {
	return template.BackendConnectors{
		c.Etcd,
		c.File,
		c.Env,
		c.Consul,
		c.Vault,
		c.Redis,
		c.Zookeeper,
		c.Mock,
	}
}

// Configuration is the representation of an config file
type Configuration struct {
	LogLevel   string `toml:"log_level"`
	LogFormat  string `toml:"log_format"`
	IncludeDir string `toml:"include_dir"`
	PidFile    string `toml:"pid_file"`
	LogFile    string `toml:"log_file"`
	Resource   []Resource
}

// Resource is the representation of an resource configuration
type Resource struct {
	Exec     template.ExecConfig
	Template []*template.Renderer
	Backend  BackendConfigs

	// defaults to the filename of the resource
	Name string
}

func readFileAndExpandEnv(path string) ([]byte, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return buf, errors.Wrap(err, "read file failed")
	}
	// expand the environment variables
	buf = []byte(os.ExpandEnv(string(buf)))
	return buf, nil
}

// NewConfiguration reads the file at `path`, expand the environment variables
// and unmarshals it to a new configuration struct.
// it returns an error if any.
func NewConfiguration(path string) (Configuration, error) {
	var c Configuration

	buf, err := readFileAndExpandEnv(path)
	if err != nil {
		return c, err
	}

	if err := toml.Unmarshal(buf, &c); err != nil {
		return c, errors.Wrapf(err, "toml unmarshal failed: %s", path)
	}

	for _, v := range c.Resource {
		if v.Name == "" {
			v.Name = filepath.Base(path)
		}
	}

	if c.IncludeDir != "" {
		files, err := ioutil.ReadDir(c.IncludeDir)
		if err != nil {
			return c, err
		}
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".toml") {
				fp := filepath.Join(c.IncludeDir, file.Name())

				log.WithFields(logrus.Fields{
					"path": fp,
				}).Info("loading resource configuration")

				buf, err := readFileAndExpandEnv(fp)
				if err != nil {
					return c, err
				}
				var r Resource
				if err := toml.Unmarshal(buf, &r); err != nil {
					return c, errors.Wrapf(err, "toml unmarshal failed: %s", fp)
				}
				// don't add empty resources
				if len(r.Template) > 0 {
					if r.Name == "" {
						r.Name = file.Name()
					}
					c.Resource = append(c.Resource, r)
				}
			}
		}
	}

	c.configureLogger()

	return c, nil
}

// configureLogger configures the global logger.
// It sets the log level and log formatting.
func (c *Configuration) configureLogger() {
	err := log.SetLevel(c.LogLevel)
	if err != nil {
		log.Error(err)
	}
	log.SetFormatter(c.LogFormat)

	err = log.SetOutput(c.LogFile)
	if err != nil {
		log.Error(err)
	}
}
