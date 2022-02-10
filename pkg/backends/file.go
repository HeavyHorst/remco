/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package backends

import (
	"github.com/HeavyHorst/easykv/file"
	berr "github.com/HeavyHorst/remco/pkg/backends/error"
	"github.com/HeavyHorst/remco/pkg/log"
	"github.com/HeavyHorst/remco/pkg/template"
	"github.com/sirupsen/logrus"
)

// FileConfig represents the config for the file backend.
type FileConfig struct {
	// The filepath to a yaml or json file containing the key-value pairs.
	// This can be a local file or a remote http/https location.
	Filepath string

	// Optional HTTP headers to append to the request if the file path is a remote http/https location.
	HTTPHeaders map[string]string
	template.Backend
}

// Connect creates a new fileClient and fills the underlying template.Backend with the file-Backend specific data.
func (c *FileConfig) Connect() (template.Backend, error) {
	if c == nil {
		return template.Backend{}, berr.ErrNilConfig
	}

	c.Backend.Name = "file"
	log.WithFields(logrus.Fields{
		"backend":  c.Backend.Name,
		"filepath": c.Filepath,
	}).Info("set file path")

	client, err := file.New(c.Filepath, file.WithHeaders(c.HTTPHeaders))
	if err != nil {
		return c.Backend, err
	}
	c.Backend.ReadWatcher = client
	return c.Backend, nil
}

// return Backend config to allow modification before connect() for onetime param or similar
func (c *FileConfig) GetBackend() *template.Backend {
	return &c.Backend
}
