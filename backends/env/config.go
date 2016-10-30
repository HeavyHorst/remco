/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package env

import (
	"github.com/HeavyHorst/easyKV/env"
	berr "github.com/HeavyHorst/remco/backends/error"
	"github.com/HeavyHorst/remco/template"
)

//Config represents the config for the env backend
type Config struct {
	template.Backend
}

// Connect creates a new envClient and fills the underlying template.Backend with the file-Backend specific data.
func (c *Config) Connect() (template.Backend, error) {
	if c == nil {
		return template.Backend{}, berr.ErrNilConfig
	}
	c.Backend.Name = "env"

	client, err := env.New()
	if err != nil {
		return c.Backend, err
	}

	c.Backend.ReadWatcher = client
	return c.Backend, nil
}
