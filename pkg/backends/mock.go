/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package backends

import (
	"github.com/HeavyHorst/easykv/mock"
	berr "github.com/HeavyHorst/remco/pkg/backends/error"
	"github.com/HeavyHorst/remco/pkg/template"
)

// MockConfig represents the config for the consul backend.
type MockConfig struct {
	Error error
	template.Backend
}

// Connect creates a new mockClient
func (c *MockConfig) Connect() (template.Backend, error) {
	if c == nil {
		return template.Backend{}, berr.ErrNilConfig
	}
	c.Backend.Name = "mock"
	client, err := mock.New(c.Error, make(map[string]string))
	if err != nil {
		return c.Backend, err
	}

	c.Backend.ReadWatcher = client

	return c.Backend, nil
}
