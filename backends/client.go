package backends

import "errors"

var ErrNilConfig = errors.New("config is nil")

type StoreClient interface {
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error)
}
