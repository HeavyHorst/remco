/*
 * This file is part of remco.
 * © 2016 The Remco Authors
 *
 * For the full copyright and license information, please view the LICENSE
 * file that was distributed with this source code.
 */

package template

import (
	"context"
	"time"

	"github.com/HeavyHorst/easykv"
	"github.com/HeavyHorst/memkv"
	berr "github.com/HeavyHorst/remco/pkg/backends/error"
	"github.com/HeavyHorst/remco/pkg/log"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

// A BackendConnector - Every backend implements this interface.
//
// If Connect is called a new connection to the underlaying kv-store will be established.
//
// Connect should also set the name and the StoreClient of the Backend. The other values of Backend will be loaded from the configuration file.
type BackendConnector interface {
	Connect() (Backend, error)
}

// BackendConnectors is a list of BackendConnectors.
// This is just a helper Type to simplify operations on many BackendConnectors.
type BackendConnectors []BackendConnector

// Backend is the representation of a template backend like etcd or consul
type Backend struct {
	easykv.ReadWatcher

	// Name is the name of the backend for example etcd or consul.
	// The name is attached to the logs.
	Name string

	// Onetime - render the config file and quit.
	Onetime bool

	// Enable/Disable watch support.
	Watch bool

	// The key-path prefix.
	Prefix string

	// The backend polling interval. Can be used as a reconcilation loop for watch or standalone.
	Interval int

	// The backend keys that the template requires to be rendered correctly.
	Keys []string

	store *memkv.Store
}

// ConnectAll connects to all configured backends.
// This method blocks until a connection to every backend has been established or the context is canceled.
func (bc BackendConnectors) ConnectAll(ctx context.Context) ([]Backend, error) {
	var backendList []Backend
	for _, config := range bc {
	retryloop:
		for {
			select {
			case <-ctx.Done():
				for _, be := range backendList {
					be.Close()
				}
				return backendList, ctx.Err()
			default:
				b, err := config.Connect()
				if err == nil {
					backendList = append(backendList, b)
				} else if err != berr.ErrNilConfig {
					log.WithFields(logrus.Fields{
						"backend": b.Name,
					}).Error(errors.Wrap(err, "connect failed"))

					//try again after 2 seconds
					time.Sleep(2 * time.Second)
					continue retryloop
				}
				break retryloop
			}
		}
	}

	return backendList, nil
}

func (s Backend) watch(ctx context.Context, processChan chan Backend, errChan chan berr.BackendError) {
	if s.Onetime {
		return
	}

	var lastIndex uint64
	keysPrefix := appendPrefix(s.Prefix, s.Keys)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			index, err := s.WatchPrefix(ctx, s.Prefix, easykv.WithKeys(keysPrefix), easykv.WithWaitIndex(lastIndex))
			if err != nil {
				if err != easykv.ErrWatchCanceled {
					errChan <- berr.BackendError{Message: err.Error(), Backend: s.Name}
					time.Sleep(2 * time.Second)
				}
				continue
			}
			processChan <- s
			lastIndex = index
		}
	}
}

func (s Backend) interval(ctx context.Context, processChan chan Backend) {
	if s.Onetime {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(s.Interval) * time.Second):
			processChan <- s
		}
	}
}
