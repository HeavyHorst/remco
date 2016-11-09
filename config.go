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

	"github.com/HeavyHorst/remco/log"
	"github.com/Sirupsen/logrus"
	"github.com/naoina/toml"
)

// configuration is the representation of an config file
type configuration struct {
	LogLevel   string `toml:"log_level"`
	LogFormat  string `toml:"log_format"`
	IncludeDir string `toml:"include_dir"`
	Resource   []resource
}

func readFileAndExpandEnv(path string) ([]byte, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return buf, err
	}
	// expand the environment variables
	buf = []byte(os.ExpandEnv(string(buf)))
	return buf, nil
}

// newConfiguration reads the file at `path`, expand the environment variables
// and unmarshals it to a new configuration struct.
// it returns an error if any.
func newConfiguration(path string) (configuration, error) {
	var c configuration

	buf, err := readFileAndExpandEnv(path)
	if err != nil {
		return c, err
	}

	if err := toml.Unmarshal(buf, &c); err != nil {
		return c, err
	}

	for _, v := range c.Resource {
		if v.Name == "" {
			v.Name = filepath.Base(path)
		}
	}

	c.configureLogger()

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
				}).Info("Loading resource configuration")

				buf, err := readFileAndExpandEnv(fp)
				if err != nil {
					return c, err
				}
				var r resource
				if err := toml.Unmarshal(buf, &r); err != nil {
					return c, err
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

	return c, nil
}

// configureLogger configures the global logger
// for example it sets the log level and log formatting
func (c *configuration) configureLogger() {
	if c.LogLevel != "" {
		err := log.SetLevel(c.LogLevel)
		if err != nil {
			log.Error(err)
		}
	}
	if c.LogFormat != "" {
		log.SetFormatter(c.LogFormat)
	}
}
