/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package config

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/HeavyHorst/remco/backends"
	backendErrors "github.com/HeavyHorst/remco/backends/error"
	"github.com/HeavyHorst/remco/executor"
	"github.com/HeavyHorst/remco/log"
	"github.com/HeavyHorst/remco/template"
	"github.com/Sirupsen/logrus"
)

// Configuration is the representation of an config file
type Configuration struct {
	LogLevel   string `toml:"log_level"`
	LogFormat  string `toml:"log_format"`
	IncludeDir string `toml:"include_dir"`
	PidFile    string `toml:"pid_file"`
	Http       string
	Resource   []Resource
}

// Resource is the representation of an resource configuration
type Resource struct {
	Exec     Exec
	Template []*template.Renderer
	Backend  backends.Config

	// defaults to the filename of the resource
	Name string
}

// Exec represents the configuration values for the exec mode
type Exec struct {
	Command      string `json:"command"`
	ReloadSignal string `toml:"reload_signal" json:"reload_signal"`
	KillSignal   string `toml:"kill_signal" json:"kill_signal"`
	KillTimeout  int    `toml:"kill_timeout" json:"kill_timeout"`
	Splay        int    `json:"splay"`
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
				var r Resource
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
func (c *Configuration) configureLogger() {
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

func (r *Resource) Init(ctx context.Context, reapLock *sync.RWMutex) (*template.Resource, error) {
	var backendList []template.Backend

	// try to connect to all backends
	// connection to all backends must succeed to continue
	for _, config := range r.Backend.GetBackends() {
	retryloop:
		for {
			select {
			case <-ctx.Done():
				for _, b := range backendList {
					b.Close()
				}
				return nil, ctx.Err()
			default:
				b, err := config.Connect()
				if err == nil {
					backendList = append(backendList, b)
				} else if err != backendErrors.ErrNilConfig {

					log.WithFields(logrus.Fields{
						"backend":  b.Name,
						"resource": r.Name,
					}).Error(err)

					// try again every 2 seconds
					time.Sleep(2 * time.Second)
					continue retryloop
				}
				// break out of the loop on success or if the backend is nil
				break retryloop
			}
		}
	}

	for _, p := range r.Template {
		p.ReapLock = reapLock
	}

	logger := log.WithFields(logrus.Fields{"resource": r.Name})
	exec := executor.New(r.Exec.Command, r.Exec.ReloadSignal, r.Exec.KillSignal, r.Exec.KillTimeout, r.Exec.Splay, logger)
	res, err := template.NewResource(backendList, r.Template, r.Name, exec)
	if err != nil {
		for _, v := range backendList {
			v.Close()
		}
	}
	return res, err
}
