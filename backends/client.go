package backends

import "errors"

// ErrNilConfig is returned if Connect is called on a nil Config
var ErrNilConfig = errors.New("config is nil")

// StoreClient defines the behaviour of the backend stores.
type StoreClient interface {
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error)
}
