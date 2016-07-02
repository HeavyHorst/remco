package backends

type Store struct {
	Name   string
	Client StoreClient
}

type StoreClient interface {
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error)
}

type BackendConfig interface {
	Connect() (Store, error)
}
